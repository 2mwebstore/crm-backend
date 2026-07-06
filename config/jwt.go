package config

func loadJWT() JWTConfig {
	return JWTConfig{
		Secret:      mustGetEnv("JWT_SECRET"),
		ExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
	}
}
