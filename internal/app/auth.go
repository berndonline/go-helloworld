package app

import (
    "crypto/subtle"
    "encoding/json"
    "github.com/dgrijalva/jwt-go"
    opentracing "github.com/opentracing/opentracing-go"
    "log"
    "net/http"
    "time"
)

var jwtKey = []byte("my_secret_key")

type Credentials struct {
    Password string `json:"password"`
    Username string `json:"username"`
}

var users = map[string]string{
    "user1": "password1",
    "user2": "password2",
}

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}

func basicAuth(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // get root span from context
        tracer := opentracing.GlobalTracer()
        span := StartSpanFromRequest("basicAuth", tracer, r)
        // basicAuth function
        realm := "Please enter your username and password"
        user, pass, ok := r.BasicAuth()
        expectedPassword := users[user]
        if !ok || subtle.ConstantTimeCompare([]byte(pass),
            []byte(expectedPassword)) != 1 {
            w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
            w.WriteHeader(401)
            w.Write([]byte("You are Unauthorized to access the application.\n"))
            log.Print("helloworld: authentication failed - " + getIPAddress(r))
            defer span.Finish()
            return
        }
        log.Print("helloworld: " + user + " login successfully - " + getIPAddress(r))
        // stop tracer and inject http infos
        defer span.Finish()
        // inject tracer into context
        Inject(span, r)
        handler(w, r)
    }
}

func jwtLogin(w http.ResponseWriter, r *http.Request) {
    var creds Credentials
    err := json.NewDecoder(r.Body).Decode(&creds)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    expectedPassword, ok := users[creds.Username]
    if !ok || expectedPassword != creds.Password {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    expirationTime := time.Now().Add(5 * time.Minute)
    claims := &Claims{
        Username: creds.Username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    http.SetCookie(w, &http.Cookie{
        Name:    "token",
        Value:   tokenString,
        Expires: expirationTime,
    })
    log.Print("helloworld: " + creds.Username + " login successfully - " + getIPAddress(r))
    w.Write([]byte("Token issued.\n"))
}

func jwtAuth(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // get root span from context
        tracer := opentracing.GlobalTracer()
        span := StartSpanFromRequest("jwtAuth", tracer, r)
        // json web token function
        c, err := r.Cookie("token")
        if err != nil {
            if err == http.ErrNoCookie {
                w.WriteHeader(http.StatusUnauthorized)
                defer span.Finish()
                return
            }
            w.WriteHeader(http.StatusBadRequest)
            defer span.Finish()
            return
        }

        tknStr := c.Value
        claims := &Claims{}
        tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })

        if err != nil {
            if err == jwt.ErrSignatureInvalid {
                w.WriteHeader(http.StatusUnauthorized)
                defer span.Finish()
                return
            }
            w.WriteHeader(http.StatusBadRequest)
            defer span.Finish()
            return
        }
        if !tkn.Valid {
            w.WriteHeader(http.StatusUnauthorized)
            defer span.Finish()
            return
        }
        // stop tracer and inject http infos
        defer span.Finish()
        // inject tracer into context
        Inject(span, r)
        handler(w, r)
    }
}

func jwtRefresh(w http.ResponseWriter, r *http.Request) {
    claims := &Claims{}
    expirationTime := time.Now().Add(5 * time.Minute)
    claims.ExpiresAt = expirationTime.Unix()
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("New token failed.\n"))
        return
    }

    http.SetCookie(w, &http.Cookie{
        Name:    "token",
        Value:   tokenString,
        Expires: expirationTime,
    })
    w.Write([]byte("Token renewed.\n"))
}

func jwtLogout(w http.ResponseWriter, r *http.Request) {
    http.SetCookie(w, &http.Cookie{
        Name:   "token",
        MaxAge: -1,
    })
    w.Write([]byte("Logged out!\n"))
}

