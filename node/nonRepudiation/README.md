# Non repudiation requirement

This module's function is to provide all functionality needed for the non repudiation protocol.

## Direct requirements

1. encryptionRequirement

## Exposed structs

All fields have get functions. The lack of set functions is intentional.

1. [**NonRepudiationRequirement**](NonRepudiationRequirementStruct.go#14) contains an rsa private key, an encryption requirement, the amount of times that the non repudiation protocol should be repeated and the fake data necessary for the non repudiation protocol.
1. [**Data**](NonRepudiationRequirementStruct.go#21) contains the fields plain password, salt and nonce. The fields are exported since it is needed for creating a JSON from the struct. The data object contains all data that is necessary to decrypt the previously received cyphertext.


## Exposed functions

1. [**EncryptMessage**](nonRepudiationStruct.go#L34) is a method for the nonRepudiationRequirement and takes the message to be encrypted. Then the struct's encryption requirement is used to encrypt the message
1. [**DecryptMessage**](nonRepudiation.go#L22) takes a data struct and the cyphertext and tries to decrypt it.
1. [**GenerateNonRepudiationRequirement**](nonRepudiation.go#L61) returns a nonRepudiationRequirement

