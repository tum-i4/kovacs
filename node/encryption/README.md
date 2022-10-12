# Encryption requirement

This module's function is to facilitate the en- and decryption by relying on the passwordRequirement module to generate password hashes which will be used as the de- and encryption key. The chosen encryption algorithm is AES GCM. The advantages include:

* It is possible to check if the decryption failed due to an invalid decryption key
* It takes a nonce which prevents the equal plaintext leading to the same ciphertext if the same key was used.

## Exposed structs

[**EncryptionRequirement**](EncryptionRequirement.go#5) contains two fields for a password requirement and a nonce. The fields have get functions. The lack of set functions is intentional.


## Exposed functions
1. [**GenerateNonce**](aesGCM.go#L13) returns a random 12 byte long nonce
2. [**EncryptAESGCM**](aesGCM.go#L24) takes the encryption key, which must be 32 byte long, a nonce for increased output entropy and the plaintext that is to be encrypted. The hex coded ciphertext is returned.
3. [**EncryptAESGCM**](aesGCM.go#L45) takes the encryption key, the nonce and the ciphertext in hex representation and returns the plaintext. The decrypted plaintext is returned.
4. [**GenerateEncryptionRequirement**](encryptionRequirement.go#10) returns an encryption requirement by requesting a password requirement from the passwordRequirement module and generating a nonce. A filled encryptionRequirement is returned.

