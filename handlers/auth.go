package handlers

import (
	"net/http"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/kflim/go-oauth/service"
	"github.com/markbates/goth/gothic"
)

func Home(c *gin.Context) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(c.Writer, gin.H{})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func SignInWithProvider(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func CallbackHandler(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	/* accessToken := user.AccessToken

	err = gothic.StoreInSession("accessToken", accessToken, c.Request, c.Writer)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return	
	}

	err = gothic.StoreInSession("userID", user.UserID, c.Request, c.Writer)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} */

	userClaims := service.UserClaims{
		UserID: user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		StandardClaims: jwt.StandardClaims{
		 IssuedAt:  time.Now().Unix(),
		 ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
		Email: user.Email,
	}

	signedAccessToken, err := service.NewAccessToken(userClaims)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	
	refreshClaims := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
	}

	signedRefreshToken, err := service.NewRefreshToken(refreshClaims)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.SetCookie("accessToken", signedAccessToken, 900, "/", "localhost", false, true)
	c.SetCookie("refreshToken", signedRefreshToken, 900, "/", "localhost", false, true)

	c.Redirect(http.StatusTemporaryRedirect, "/success")
}

func Success(c *gin.Context) {
	/* accessToken, err := gothic.GetFromSession("accessToken", c.Request)

	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := template.ParseFiles("templates/chat.html")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(c.Writer, gin.H{})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} */

	userClaims := service.ParseAccessToken(c.Request.CookiesNamed("accessToken")[0].Value)
	refreshClaims := service.ParseRefreshToken(c.Request.CookiesNamed("refreshToken")[0].Value)

	if refreshClaims.Valid() != nil || userClaims.StandardClaims.Valid() != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	tmpl, err := template.ParseFiles("templates/chat.html")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(c.Writer, gin.H{})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func RetryLogin(c *gin.Context) {
	tmpl, err := template.ParseFiles("templates/retry-login.html")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(c.Writer, gin.H{})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
			accessToken := c.Request.CookiesNamed("accessToken")[0].Value // assuming token is stored in a cookie
			if accessToken == "" {
					// Token is missing or invalid
					c.Redirect(http.StatusSeeOther, "/retry-login")
					c.Abort()
					return
			}

			// Parse and validate the token (adjust according to your JWT implementation)
			userClaims := service.ParseAccessToken(accessToken) // Assuming you have this service
			if userClaims == nil {
					// Token is invalid
					c.Redirect(http.StatusSeeOther, "/retry-login")
					c.Abort()
					return
			}

			refreshToken := c.Request.CookiesNamed("refreshToken")[0].Value
			if refreshToken == "" {
				// Token is missing or invalid
				c.Redirect(http.StatusSeeOther, "/retry-login")
				c.Abort()
				return
			}

			refreshClaims := service.ParseRefreshToken(refreshToken)
			if refreshClaims == nil {
				// Token is invalid
				c.Redirect(http.StatusSeeOther, "/retry-login")
				c.Abort()
				return
			}

			// Token is valid, continue with the request
			c.Next()
	}
}

