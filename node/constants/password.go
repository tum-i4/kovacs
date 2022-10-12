package constants

const (
	NonceLength          = 12
	PasswordPlainLength  = 32
	PasswordHashedLength = 32
	SaltLength           = 22 // Since the salt is base64 encoded

	// HashDifficulty is set to 16 due to measurements from generatePasswordTimer and time-out time of 3 seconds.
	HashDifficulty = 16
	// RequirementListLength shouldn't be too high to avoid large wait time at peer start
	// Creating a password takes ~3 seconds, thus filling the list takes about (3s*20)/60 ~ 1 minute.
	RequirementListLength = 2
)
