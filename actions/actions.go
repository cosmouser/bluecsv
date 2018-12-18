package actions

import (
	"errors"
	"html/template"
	"net/http"
	"sync"

	"github.com/cosmouser/bluecsv/config"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type templateInfo struct {
	LoggedIn    string
	ExternalURL string
	AdminEmail  string
	Error       error
}

type sessionStore struct {
	sync.RWMutex
	cache map[string]string
}

// FaqHandler provides the FAQ page
func FaqHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := session(w, r)
	templateFile := *config.TmplPath + "faq.tmpl"
	faqTmpl := template.Must(template.ParseFiles(templateFile))
	log.WithFields(log.Fields{
		"userID": userID,
		"page":   "Faq",
	}).Info()

	info := &templateInfo{
		LoggedIn:    userID,
		ExternalURL: config.C.ExternalURL,
		AdminEmail:  config.C.AdminEmail,
	}
	faqTmpl.Execute(w, info)
	return
}

// HomeHandler provides the form or the login page depending on the session
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := session(w, r)
	defer r.Body.Close()
	privateTemplate := *config.TmplPath + "form.tmpl"
	publicTemplate := *config.TmplPath + "login.tmpl"
	var homeTmpl *template.Template
	if err != nil {
		homeTmpl = template.Must(template.ParseFiles(publicTemplate))

	} else {
		homeTmpl = template.Must(template.ParseFiles(privateTemplate))
	}
	log.WithFields(log.Fields{
		"userID": userID,
		"page":   "Home",
	}).Info()
	info := &templateInfo{
		LoggedIn:    userID,
		ExternalURL: config.C.ExternalURL,
		AdminEmail:  config.C.AdminEmail,
	}
	homeTmpl.Execute(w, info)
}

// Authenticate handles logins to use the web app
func Authenticate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	defer r.Body.Close()
	if userID, ok := config.Credentials[r.PostFormValue("key")]; ok {
		uUID := store.put(userID)
		cookie := &http.Cookie{
			Name:     "_session",
			Value:    uUID,
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		log.WithFields(log.Fields{
			"key": r.PostFormValue("key"),
		}).Info("Login Successful")
		http.Redirect(w, r, config.C.ExternalURL+"/", 302)
	} else {
		log.WithFields(log.Fields{
			"key": r.PostFormValue("key"),
		}).Warn("Login Failed")
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalURL: config.C.ExternalURL,
			Error:       errors.New("Invalid Login"),
		}
		errorTmpl.Execute(w, info)
		return
	}
}

// Logout clears the session
func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("_session")
	if err == nil {
		_, ok := store.check(cookie.Value)
		if !ok {
			http.Redirect(w, r, config.C.ExternalURL+"/", 302)
		}
		wipedCookie := &http.Cookie{
			Name:     "_session",
			Value:    "logged_out",
			HttpOnly: true,
		}
		http.SetCookie(w, wipedCookie)
	}
	http.Redirect(w, r, config.C.ExternalURL+"/", 302)
}

func (store *sessionStore) check(uuid string) (userID string, ok bool) {
	store.RLock()
	userID, ok = store.cache[uuid]
	store.RUnlock()
	return userID, ok
}
func (store *sessionStore) put(userID string) (uUID string) {
	uUID = uuid.Must(uuid.NewV4()).String()
	store.Lock()
	log.WithFields(log.Fields{
		"uuid":   uUID,
		"userID": userID,
	}).Info("New Session")
	store.cache[uUID] = userID
	store.Unlock()
	return uUID
}
func newStore() *sessionStore {
	return &sessionStore{
		cache: make(map[string]string),
	}
}
func session(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie("_session")
	var userID string
	if err == nil {
		id, ok := store.check(cookie.Value)
		if !ok {
			err = errors.New("Invalid Session")
		}
		userID = id
	}
	return userID, err
}

var errorTmpl *template.Template
var store *sessionStore

func init() {
	var err error
	errorTmpl, err = template.ParseFiles(*config.TmplPath + "error.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	store = newStore()
}
