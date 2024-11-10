func SetupFactrakTest(t *testing.T) (*testify.Assertions, *gorm.DB, *gin.Engine) {
	// Create test environment using factrak scopes
	env := utils.SetupTest(t, auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	// Set up the factrak router
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())

	return env.Assert, env.DB, env.Router
}