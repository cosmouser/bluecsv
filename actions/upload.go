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

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := session(w, r)
	defer r.Body.Close()
	if r.Method == "GET" {
		http.Redirect(w, r, config.C.ExternalUrl+"/", 302)
	}
	if err != nil || r.Method != "POST" {
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalUrl: config.C.ExternalUrl,
			Error:       errors.New("Invalid Operation"),
		}
		ErrorTmpl.Execute(w, info)
		return
	}
	formReader, err := r.MultipartReader()
	if err != nil {
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalUrl: config.C.ExternalUrl,
			Error:       err,
		}
		ErrorTmpl.Execute(w, info)
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
				ExternalUrl: config.C.ExternalUrl,
				Error:       err,
			}
			ErrorTmpl.Execute(w, info)
			return
		}
		n, err := part.Read(columnBuffer)
		if err != nil {
			if err != io.EOF {
				info := &templateInfo{
					LoggedIn:    userID,
					ExternalUrl: config.C.ExternalUrl,
					Error:       err,
				}
				ErrorTmpl.Execute(w, info)
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
			ExternalUrl: config.C.ExternalUrl,
			Error:       err,
		}
		ErrorTmpl.Execute(w, info)
		return
	}
	if filename := part.FileName(); len(filename) < 4 {
		log.WithFields(log.Fields{
			"filename": filename,
		}).Error("Invalid File Format")
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalUrl: config.C.ExternalUrl,
			Error:       errors.New("Please upload a CSV file with org emails in the first column"),
		}
		ErrorTmpl.Execute(w, info)
		return
	}
	if filename := part.FileName(); filename[len(filename)-3:] != "csv" {
		log.WithFields(log.Fields{
			"filename": filename,
			"fileExt":  filename[len(filename)-3:],
		}).Error("Invalid File Format")
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalUrl: config.C.ExternalUrl,
			Error:       errors.New("Invalid file format. Please make sure to export your data in csv format before uploading."),
		}
		ErrorTmpl.Execute(w, info)
		return
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf(
			"attachment; filename=%v_out.csv",
			filename[:len(filename)-4],
		))
		w.Header().Set("Content-Type", "text/csv")
	}
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
			ExternalUrl: config.C.ExternalUrl,
			Error:       errors.New("Header row has too few columns. Please see FAQ for details."),
		}
		ErrorTmpl.Execute(w, info)
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
					ExternalUrl: config.C.ExternalUrl,
					Error:       err,
				}
				ErrorTmpl.Execute(w, info)
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
