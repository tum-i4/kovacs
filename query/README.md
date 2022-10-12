# Purpose

This program allows for easy access to the blockchain storage and an SQLite database should it exist. It returns the usage log that is associated with the passed pseudonyms.

# Example

The schema is as follows: ```./query [flags] [location of non-repudiation logs]```. An example would be: ```./query -all ../requester/storage```, which would return all logs associated with the data consumer's pseudonyms.
