# sqlgrl
> pronounced 'squirrel girl'

an oracle sql to microsoft sql conversion tool for DDL table definitions / exports

## implemented
generic/generic.go - common table definitions structures and helper functions
generic/lexer.go - functions for reading string content into tokens
oracle/tokenizer.go - implementation of tokens for oracle sql (following test cases not specs)
main.go - crawls a directory or individual file as first arg and runs conversion over .sql files

## todo
oracle/parser.go - convert tokens to common table structs
tsql/serializer.go - convert common table structs to t-sql format