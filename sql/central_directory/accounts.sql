DROP DATABASE IF EXISTS accounts;
CREATE DATABASE accounts;
USE accounts;

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts(
	id BINARY(16) NOT NULL,
    name VARCHAR(20) NOT NULL,
    created DATETIME NOT NULL,
    region CHAR(3) NOT NULL,
    shard MEDIUMINT NOT NULL DEFAULT -1,
    isUser BOOL NOT NULL,
    PRIMARY KEY (name),
    UNIQUE INDEX (id)
);

DROP TABLE IF EXISTS users;
CREATE TABLE users(
	id       BINARY(16) NOT NULL,
	email    VARCHAR(250) NOT NULL,
	newEmail VARCHAR(250) NULL,
	ActivationCode VARCHAR(100) NULL,
	Activated DATETIME NULL,
	NewEmailConfirmationCode VARCHAR(100) NULL,
	ResetPwdCode VARCHAR(100) NULL,
    PRIMARY KEY (id),
    UNIQUE INDEX (email),
    FOREIGN KEY (id) REFERENCES accounts (id)
);

DROP USER IF EXISTS 'task_center_central_directory_api'@'%';
CREATE USER 'task_center_central_directory_api'@'%' IDENTIFIED BY 'T@sk-C3n-T3r'; 
GRANT SELECT (id, salt, pwd, n, r, p, keyLen) ON pwds.pwds TO 'task_center_central_directory_api'@'%';