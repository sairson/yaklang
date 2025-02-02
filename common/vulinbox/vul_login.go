package vulinbox

import (
	_ "embed"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/dgrijalva/jwt-go"
	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/mutate"
	"github.com/yaklang/yaklang/common/utils"
	"net/http"
	"strings"
)

//go:embed vul_login_jwt_login.html
var jwtLoginPage []byte

//go:embed vul_login_jwt_profile.html
var jwtLoginProfilePage []byte

//go:embed vul_login_login_setjwt.html
var jwtLoginProfileSetJWTPage []byte

func (s *VulinServer) registerLoginRoute() {
	var r = s.router

	key := utils.RandStringBytes(20)
	var keyF jwt.Keyfunc = func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}

	r.HandleFunc("/jwt/login/profile", func(writer http.ResponseWriter, request *http.Request) {
		authToken := request.Header.Get("Authorization")
		if authToken != "" {
			first, _, _ := strings.Cut(" ", authToken)
			if first != "Bearer" {
				writer.WriteHeader(401)
				writer.Write([]byte("invalid auth token, use Bearer schema"))
				return
			}
			token, err := jwt.Parse(authToken, keyF)
			if err != nil {
				writer.WriteHeader(401)
				writer.Write([]byte("invalid auth token"))
				return
			}
			if !token.Valid {
				writer.WriteHeader(401)
				writer.Write([]byte("invalid auth token"))
				return
			}

			writer.Header().Set("Content-Type", "text/html")
			name := utils.MapGetString(token.Header, "username")
			users, err := s.database.GetUserByUsernameUnsafe(name)
			if err != nil {
				writer.WriteHeader(500)
				writer.Write([]byte("internal error, cannot found user: " + name))
				return
			}

			jsonRaw, err := json.Marshal(funk.Map(users, func(u *VulinUser) map[string]any {
				var a = make(map[string]any)
				a["username"] = u.Username
				a["id"] = u.ID
				a["age"] = u.Age
				a["updated_at"] = u.UpdatedAt.String()
				a["created_at"] = u.CreatedAt.String()
				return a
			}))
			if err != nil {
				writer.WriteHeader(500)
				writer.Write([]byte("internal error, cannot found user: " + name + " \n json.Marshal failed: " + err.Error()))
				return
			}

			data, _ := mutate.FuzzTagExec(jwtLoginProfilePage, mutate.Fuzz_WithParams(map[string]any{
				"jsonRaw": string(jsonRaw),
			}))
			writer.Write([]byte(data[0]))
			return
		}
		writer.WriteHeader(401)
		writer.Write([]byte("invalid auth token"))
		return
	})
	r.HandleFunc("/jwt/login", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "GET" {
			// 不存在登录信息
			writer.Header().Set("Content-Type", "text/html")
			writer.Write(jwtLoginPage)
			return
		}

		if request.Method == "POST" {
			// 登录
			username := request.FormValue("username")
			password := request.FormValue("password")
			if username == "" || password == "" {
				writer.WriteHeader(400)
				writer.Write([]byte("username or password cannot be empty"))
				return
			}

			users, err := s.database.GetUserByUsernameUnsafe(username)
			if err != nil {
				writer.WriteHeader(500)
				writer.Write([]byte("internal error, cannot found user: " + username))
				return
			}
			if len(users) == 0 {
				writer.WriteHeader(400)
				writer.Write([]byte("username or password incorrect"))
				return
			}

			user := users[0]
			if user.Password != password {
				writer.WriteHeader(400)
				writer.Write([]byte("username or password incorrect"))
				return
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"username": user.Username,
			})
			token.Header["kid"] = user.ID
			token.Header["username"] = user.Username
			token.Header["age"] = user.Age

			tokenString, err := token.SignedString([]byte(key))
			if err != nil {
				writer.WriteHeader(500)
				writer.Write([]byte("internal error, cannot sign token: " + err.Error() + "\n " + spew.Sdump(key)))
				return
			}

			writer.Header().Set("Content-Type", "text/html")
			jsonBytes := []byte(`{"token": "` + string(tokenString) + `"}`)
			data, _ := mutate.FuzzTagExec(jwtLoginProfileSetJWTPage, mutate.Fuzz_WithParams(map[string]any{
				"jsonRaw": string(jsonBytes),
			}))
			writer.Write([]byte(data[0]))
			return
		}

		writer.WriteHeader(405)
		writer.Write([]byte("method not allowed"))
	})
}
