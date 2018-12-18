package actions

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cosmouser/bluecsv/config"
	log "github.com/sirupsen/logrus"
)

// UploadHandler receives the uploaded CSV and returns the output CSV
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := session(w, r)
	defer r.Body.Close()
	if r.Method == "GET" {
		http.Redirect(w, r, config.C.ExternalURL+"/", 302)
	}
	if err != nil || r.Method != "POST" {
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalURL: config.C.ExternalURL,
			Error:       errors.New("Invalid Operation"),
		}
		errorTmpl.Execute(w, info)
		return
	}
	formReader, err := r.MultipartReader()
	if err != nil {
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalURL: config.C.ExternalURL,
			Error:       err,
		}
		errorTmpl.Execute(w, info)
		return
	}
	// Read values of col1 -- col6
	numCols := 6
	attributes := []string{}
	for i := 0; i < numCols; i++ {
		columnBuffer := make([]byte, 32)
		part, err := formReader.NextPart()
		if err != nil {
			info := &templateInfo{
				LoggedIn:    userID,
				ExternalURL: config.C.ExternalURL,
				Error:       err,
			}
			errorTmpl.Execute(w, info)
			return
		}
		n, err := part.Read(columnBuffer)
		if err != nil {
			if err != io.EOF {
				info := &templateInfo{
					LoggedIn:    userID,
					ExternalURL: config.C.ExternalURL,
					Error:       err,
				}
				errorTmpl.Execute(w, info)
				return
			}
		}
		if n > 0 {
			attribute := string(columnBuffer[:n])
			attributes = append(attributes, attribute)
		}
	}
	// read the csv
	part, err := formReader.NextPart()
	if err != nil {
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalURL: config.C.ExternalURL,
			Error:       err,
		}
		errorTmpl.Execute(w, info)
		return
	}
	filename := part.FileName()
	if len(filename) < 4 {
		log.WithFields(log.Fields{
			"filename": filename,
		}).Error("invalid file format")
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalURL: config.C.ExternalURL,
			Error:       errors.New("please upload a CSV file with org emails in the first column"),
		}
		errorTmpl.Execute(w, info)
		return
	}
	if filename[len(filename)-3:] != "csv" {
		log.WithFields(log.Fields{
			"filename": filename,
			"fileExt":  filename[len(filename)-3:],
		}).Error("invalid file format")
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalURL: config.C.ExternalURL,
			Error:       errors.New("invalid file format. please make sure to export your data in csv format before uploading"),
		}
		errorTmpl.Execute(w, info)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf(
		"attachment; filename=%v_out.csv",
		filename[:len(filename)-4],
	))
	w.Header().Set("Content-Type", "text/csv")

	log.WithFields(log.Fields{
		"userID":     userID,
		"attributes": attributes,
	}).Info("Processing Upload")
	var emptyLines, numSearches int
	csvReader := csv.NewReader(part)
	records, err := csvReader.ReadAll()
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalURL: config.C.ExternalURL,
			Error:       errors.New("header row has too few columns. please see FAQ for details"),
		}
		errorTmpl.Execute(w, info)
		return
	}
	log.WithFields(log.Fields{
		"userID":     userID,
		"attributes": attributes,
		"rows":       len(records),
	}).Info("CSV Uploaded")
	csvWriter := csv.NewWriter(w)
	for _, record := range records {
		if len(record) > 0 {
			if numSearches > 0 && numSearches%64 == 0 {
				time.Sleep(1 * time.Second)
			}
			newData, err := GetLdapValues(record[0], attributes)
			record = append(record, newData...)
			if err != nil {
				w.Header().Set("Content-Type", "text/html")
				info := &templateInfo{
					LoggedIn:    userID,
					ExternalURL: config.C.ExternalURL,
					Error:       err,
				}
				errorTmpl.Execute(w, info)
				return
			}
			emptyLines = 0
			numSearches++
		} else {
			emptyLines++
		}
		if emptyLines > 3 {
			break
		}
		err = csvWriter.Write(record)
		if err != nil {
			log.Error(err)
		}
	}
	csvWriter.Flush()
}
