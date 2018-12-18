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
	ExternalUrl string
	AdminEmail  string
	Error       error
}
type SessionStore struct {
	sync.RWMutex
	cache map[string]string
}

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
		ExternalUrl: config.C.ExternalUrl,
		AdminEmail:  config.C.AdminEmail,
	}
	faqTmpl.Execute(w, info)
	return
}
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
		ExternalUrl: config.C.ExternalUrl,
		AdminEmail:  config.C.AdminEmail,
	}
	homeTmpl.Execute(w, info)
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	defer r.Body.Close()
	if userID, ok := config.Credentials[r.PostFormValue("key")]; ok {
		uUid := Store.Put(userID)
		cookie := &http.Cookie{
			Name:     "_session",
			Value:    uUid,
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		log.WithFields(log.Fields{
			"key": r.PostFormValue("key"),
		}).Info("Login Successful")
		http.Redirect(w, r, config.C.ExternalUrl+"/", 302)
	} else {
		log.WithFields(log.Fields{
			"key": r.PostFormValue("key"),
		}).Warn("Login Failed")
		info := &templateInfo{
			LoggedIn:    userID,
			ExternalUrl: config.C.ExternalUrl,
			Error:       errors.New("Invalid Login"),
		}
		ErrorTmpl.Execute(w, info)
		return
	}
}
func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("_session")
	if err == nil {
		_, ok := Store.Check(cookie.Value)
		if !ok {
			http.Redirect(w, r, config.C.ExternalUrl+"/", 302)
		}
		wipedCookie := &http.Cookie{
			Name:     "_session",
			Value:    "logged_out",
			HttpOnly: true,
		}
		http.SetCookie(w, wipedCookie)
	}
	http.Redirect(w, r, config.C.ExternalUrl+"/", 302)
}
func (store *SessionStore) Check(uuid string) (userID string, ok bool) {
	store.RLock()
	userID, ok = store.cache[uuid]
	store.RUnlock()
	return userID, ok
}
func (store *SessionStore) Put(userID string) (uUid string) {
	uUid = uuid.Must(uuid.NewV4()).String()
	store.Lock()
	log.WithFields(log.Fields{
		"uuid":   uUid,
		"userID": userID,
	}).Info("New Session")
	store.cache[uUid] = userID
	store.Unlock()
	return uUid
}
func NewStore() *SessionStore {
	return &SessionStore{
		cache: make(map[string]string),
	}
}
func session(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie("_session")
	var userID string
	if err == nil {
		id, ok := Store.Check(cookie.Value)
		if !ok {
			err = errors.New("Invalid Session")
		}
		userID = id
	}
	return userID, err
}

var ErrorTmpl *template.Template
var Store *SessionStore

func init() {
	var err error
	ErrorTmpl, err = template.ParseFiles(*config.TmplPath + "error.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	Store = NewStore()
}
