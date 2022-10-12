Adapted the [offical Go bcrypt implementation](https://github.com/golang/crypto/blob/c07d793c2f9a/bcrypt/bcrypt.go) in
order to prevent the usage of random salts which allows for reproducibility. This is achieved by two functions:
1. Using [GeneratePasswordReturnSalt](bcrypt.go) takes the password and creates a random salt which is then used to calculates the hash. Afterwards the hash and salt are returned. This function is for the data owner.
1. Using [GeneratePasswordFromSalt](bcrypt.go) takes a salt and a password and calculates the hash from that. This functino is for the data consumer.
