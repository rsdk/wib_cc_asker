package hello

import (
	"html/template"
	//"io/ioutil"
	"fmt"
	"net/http"
	//"net/url"
	"strconv"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

var (
	backendTemplate = template.Must(template.ParseFiles("backend.html"))
	addqTemplate    = template.Must(template.ParseFiles("addq.html"))
	askTemplate     = template.Must(template.ParseFiles("ask.html"))
	activeTemplate  = template.Must(template.ParseFiles("active.html"))
	resultTemplate  = template.Must(template.ParseFiles("results.html"))
	dankeTemplate   = template.Must(template.ParseFiles("danke.html"))
)

type Question struct {
	Question string
	Date     time.Time
	Choices  []string
	Active   bool
}

type Answer struct {
	Date   time.Time
	Count  []int
	Active bool
	//QKey  datastore.Key
}

type Uurl struct {
	Userid string
	Uurl   string
}

func init() {
	http.HandleFunc("/", ask)
	http.HandleFunc("/backend", backend)
	http.HandleFunc("/count", count)
	http.HandleFunc("/danke", danke)
	http.HandleFunc("/backend/active", active)
	http.HandleFunc("/backend/deactivate", deactivate)
	//http.HandleFunc("/backend/results", results)
	http.HandleFunc("/backend/addq", backend_addq)
	http.HandleFunc("/backend/saveq", backend_saveq)
	http.HandleFunc("/backend/delq", backend_delq)
	http.HandleFunc("/backend/save_new_url", be_save_new_url)
	//http.HandleFunc("/", landing)
}

func user_rootkey(ctx appengine.Context, userid string) *datastore.Key {
	return datastore.NewKey(ctx, "Question", userid, 0, nil)
}

func url_rootkey(ctx appengine.Context) *datastore.Key {
	return datastore.NewKey(ctx, "URL", "URL", 0, nil)
}

func count(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	key_str := r.FormValue("key")
	nr_str := r.FormValue("idnr")
	nr, err := strconv.ParseInt(nr_str, 10, 64)
	ctx.Debugf("%v", nr_str)
	key, err := datastore.DecodeKey(key_str)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var answer []Answer
	qu := datastore.NewQuery("Answer").Ancestor(key).Filter("Active = ", true).Limit(1)
	keys, err := qu.GetAll(ctx, &answer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(answer) == 0 {
		fmt.Fprint(w, "no Answer")
		return
	}

	answer[0].Count[nr-1] += 1
	ctx.Debugf("%v", answer[0].Count[nr-1])

	_, err = datastore.Put(ctx, keys[0], &answer[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/danke")
	w.WriteHeader(http.StatusFound)
	return

}

func active(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	key_str := r.FormValue("key")
	key, err := datastore.DecodeKey(key_str)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e := new(Question)
	err = datastore.Get(ctx, key, e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e.Active = true
	_, err = datastore.Put(ctx, key, e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	url_rk := url_rootkey(ctx)
	var uid []Uurl
	qu := datastore.NewQuery("URL").Ancestor(url_rk).Filter("Userid = ", u.ID).Limit(1)
	_, err = qu.GetAll(ctx, &uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(uid) == 0 {
		fmt.Fprint(w, "no URL")
		return
	}
	answer := new(Answer)
	answer.Date = time.Now()
	arr := make([]int, 5)
	ctx.Debugf("%v", arr)
	answer.Count = arr
	answer.Active = true
	//answer.QKey = *key
	//root_key := datastore.NewKey(ctx, "Answer", u.ID, 0, nil)
	na_key := datastore.NewIncompleteKey(ctx, "Answer", key)
	a_key, err := datastore.Put(ctx, na_key, answer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//ctx.Debugf(r.URL.Host)

	data := struct {
		Url      string
		Key_a    string
		Key      string
		Myurl    string
		Myurlraw string
	}{
		uid[0].Uurl,
		a_key.Encode(),
		key_str,
		r.URL.Host + uid[0].Uurl, //url.QueryEscape(r.URL.Host + uid[0].Uurl),
		r.URL.Host + uid[0].Uurl,
	}

	err = activeTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func deactivate(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	//u := user.Current(ctx)
	key_str := r.FormValue("key")
	keya_str := r.FormValue("keya")
	key, err := datastore.DecodeKey(key_str)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	keya, err := datastore.DecodeKey(keya_str)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e := new(Question)
	err = datastore.Get(ctx, key, e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e.Active = false
	_, err = datastore.Put(ctx, key, e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a := new(Answer)
	err = datastore.Get(ctx, keya, a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Active = false
	_, err = datastore.Put(ctx, keya, a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type As struct {
		A string
		P int
	}

	as := make([]As, 5)
	for i := 0; i < 5; i++ {
		as[i].A = e.Choices[i]
		as[i].P = a.Count[i]
		//ctx.Debugf("%v", as[i].Nr)
	}

	data := struct {
		Q   string
		Aws []As
	}{
		e.Question,
		as,
	}

	err = resultTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func ask(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	url_rk := url_rootkey(ctx)
	var uurl []Uurl
	q := datastore.NewQuery("URL").Ancestor(url_rk).Filter("Uurl = ", r.URL.Path).Limit(1)
	q.GetAll(ctx, &uurl)

	if len(uurl) == 0 {
		fmt.Fprint(w, "nothing here")
		return
	}
	user_rk := user_rootkey(ctx, uurl[0].Userid)
	var question []Question
	qu := datastore.NewQuery("Question").Ancestor(user_rk).Filter("Active = ", true).Limit(1)
	keys, err := qu.GetAll(ctx, &question)
	if len(question) == 0 {
		fmt.Fprint(w, "no active question")
		return
	}

	type As struct {
		Ans string
		Nr  string
	}

	as := make([]As, 0, 5)
	var a As
	for i := 0; i < len(question[0].Choices); i++ {

		a.Ans = question[0].Choices[i]
		a.Nr = strconv.Itoa(i + 1)
		if len(a.Ans) > 0 {
			as = append(as, a)
			ctx.Debugf("%v", len(as))
		}
	}

	data := struct {
		Question string
		Choices  []As
		Key      string
	}{
		question[0].Question,
		as,
		keys[0].Encode(),
	}

	err = askTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func backend_saveq(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)

	rootkey := user_rootkey(ctx, u.ID)

	if r.FormValue("q") != "" {
		key := datastore.NewIncompleteKey(ctx, "Question", rootkey)
		var q Question
		q.Question = r.FormValue("q")
		q.Date = time.Now()
		a := make([]string, 5)
		for i := 0; i < 5; i++ {
			a[i] = r.FormValue("c" + strconv.Itoa(i))

		}
		q.Active = false
		q.Choices = a

		_, err := datastore.Put(ctx, key, &q)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Location", "/backend")
	w.WriteHeader(http.StatusFound)
	return
}

func danke(w http.ResponseWriter, r *http.Request) {

	data := struct {
		danke string
	}{
		"test",
	}

	err := dankeTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func backend_addq(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	//count, _ := strconv.Atoi(r.FormValue("count"))
	//rootkey := user_rootkey(ctx, u.ID)

	data := struct {
		User string
		Id   string
	}{
		u.Email,
		u.ID,
	}

	err := addqTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func backend_delq(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	key_str := r.FormValue("key")
	key, err := datastore.DecodeKey(key_str)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = datastore.Delete(ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", "/backend")
	w.WriteHeader(http.StatusFound)
	return
}

func be_save_new_url(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	url_rk := url_rootkey(ctx)
	ctx.Debugf(r.FormValue("url"))
	var uid []Uurl
	qu := datastore.NewQuery("URL").Ancestor(url_rk).Filter("Uurl = ", r.FormValue("url")).Limit(1)
	_, err := qu.GetAll(ctx, &uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(uid) == 0 {
		key := datastore.NewIncompleteKey(ctx, "URL", url_rk)
		var uid2 Uurl
		uid2.Userid = u.ID
		uid2.Uurl = "/" + r.FormValue("url")
		_, err := datastore.Put(ctx, key, &uid2)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", "/backend")
		w.WriteHeader(http.StatusFound)
		return
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

func backend(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil {
		url, err := user.LoginURL(ctx, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}
	
	// Prüfen ob der User schon eine URL fuer Fragen hat
	url_rk := url_rootkey(ctx)
	var uid []Uurl  // Array aus Uurl structs deklarieren
	qu := datastore.NewQuery("URL").Ancestor(url_rk).Filter("Userid = ", u.ID).Limit(1)
	keys, err := qu.GetAll(ctx, &uid) // Die Adresse des Uurl structs uebergeben
	
	// Wenn Fehler beim Daten holen, 500er Fehler anzeigen
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Wenn es zu diesem User einen Eintrag gibt, die URL in uid_out speichern
	var uid_out string
	if len(uid) > 0 {
		uid_out = uid[0].Uurl
	}

	// Alle Fragen zu diesem User aus dem Datastore holen
	rootkey := user_rootkey(ctx, u.ID)
	q := datastore.NewQuery("Question").Ancestor(rootkey).Order("-Date")
	var questions []Question
	keys, err = q.GetAll(ctx, &questions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	
	type Qs struct {
		Question string
		Date     string
		Key      string
		Active   bool
	}

	
	qs := make([]Qs, len(keys)) // Slice vom struct Qs erzeugen
	for i := 0; i < len(keys); i++ {
		qs[i].Question = questions[i].Question
		qs[i].Date = questions[i].Date.Format("2006-01-02 15:04")
		qs[i].Key = keys[i].Encode()
		qs[i].Active = questions[i].Active
	}

	lo_url, _ := user.LogoutURL(ctx, "/backend")
	//rootkey := user_rootkey(ctx, u.ID)

	// Struct fuer die Uebergabe an die Templating Engine erzeugen
	data := struct {
		User      string
		Questions *[]Qs
		Lo_url    string
		Uurl      string
	}{
		u.Email,
		&qs,     // Die Adresse des Qs Slice
		lo_url,
		uid_out,
	}

	err = backendTemplate.Execute(w, data)   // Template rendern
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
