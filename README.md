# Router

simple map based go http router

    r := router.InitRouter()
    r.AddRoute("get", "/", someview)
    http.ListenAndServe(port, r)


