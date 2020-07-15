package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"

	"sibsiu.ru/controllers"
	"sibsiu.ru/middleware"
	"sibsiu.ru/models"
)

func main() {

	// flags' initialization

	boolPtr := flag.Bool("prod", false, "Provide this flag "+
		"in production. This ensures that a .config file is "+
		"provided before the application starts.")
	flag.Parse()

	// the app's config's initialization

	cfg := LoadConfig(*boolPtr)
	dbCfg := cfg.Database

	// creating services

	services, err := models.NewServices(
		models.WithGorm(dbCfg.Dialect(), dbCfg.ConnectionInfo()),
		// Only log when not in prod
		//models.WithLogMode(!cfg.IsProd()),
		models.WithUser(cfg.Pepper, cfg.HMACKey),
		models.WithScore(),
		models.WithGroup(),
		models.WithSkip(),
		models.WithTesting(),
	)
	if err != nil {
		panic(err)
	}
	defer services.Close()

	fmt.Println("Successfully connected!")

	r := mux.NewRouter()

	// initializing controllers

	staticC := controllers.NewStatic(services.Score)
	usersC := controllers.NewUsers(services.User)
	scoresC := controllers.NewScores(services.Score)
	groupsC := controllers.NewGroups(services.Group)
	skipC := controllers.NewSkips(services.Skip)

	// creating middleware

	userMw := middleware.User{
		UserService: services.User,
	}
	requireUserMw := middleware.RequireUser{}
	requireClassesMw := middleware.RequireClasses{}

	// todo почему-то не робит csrf
	//b, err := rand.Bytes(32)
	//must(err)
	//csrfMw := csrf.Protect(b, csrf.Secure(cfg.IsProd()))

	// routing
	r.HandleFunc("/api/filter/skips", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(skipC.Filter, "Студент", "Староста"))).Methods("GET")
	r.HandleFunc("/api/filter/skips/edit", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(skipC.FilterEdit, "Староста"))).Methods("GET")
	r.HandleFunc("/skips/edit", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(skipC.Edit, "Староста", "Студент"))).Methods("GET")
	r.HandleFunc("/scores/report", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(scoresC.Report, "Староста"))).Methods("GET")
	r.HandleFunc("/api/skips/update", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(skipC.Update, "Староста", "Студент"))).Methods("POST")
	r.HandleFunc("/skips", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(skipC.Show, "Староста", "Студент"))).Methods("GET")
	r.HandleFunc("/api/groups/statuses/update", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(groupsC.UpdateStatuses, "Староста"))).Methods("POST")
	r.HandleFunc("/groups/statuses", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(groupsC.ShowStatuses, "Староста"))).Methods("GET")
	r.HandleFunc("/api/scores/update", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(scoresC.Update, "Староста"))).Methods("POST")
	r.HandleFunc("/scores", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(scoresC.Show, "Студент", "Староста"))).Methods("GET")
	r.HandleFunc("/api/filter/scores", requireUserMw.ApplyFn(requireClassesMw.ApplyFn(scoresC.Filter, "Студент", "Староста"))).Methods("GET")

	// User + static routes
	r.HandleFunc("/", staticC.Root).Methods("GET")
	//r.HandleFunc("/signup", usersC.New).Methods("GET")
	//r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	// NOTE: We are using the Handle function, not HandleFunc
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.Handle("/logout", requireUserMw.ApplyFn(usersC.Logout)).Methods("POST")

	// Image routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Assets (css, js)
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	//csrfMw(userMw.Apply(r))
	must(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), userMw.Apply(r)))
}

// A helper function that panics on any error
func must(err error) {
	if err != nil {
		panic(err)
	}
}
