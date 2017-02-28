DROP DATABASE IF EXISTS accounts;
CREATE DATABASE accounts;
USE accounts;

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts(
	id BINARY(16) NOT NULL,
    name VARCHAR(20) NOT NULL,
    created DATETIME NOT NULL,
    region CHAR(3) NOT NULL,
    newRegion CHAR(3) NULL,
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
	activationCode VARCHAR(100) NULL,
	activated DATETIME NULL,
	newEmailConfirmationCode VARCHAR(100) NULL,
	resetPwdCode VARCHAR(100) NULL,
    PRIMARY KEY (id),
    UNIQUE INDEX (email),
    FOREIGN KEY (id) REFERENCES accounts (id)
);

DROP TABLE IF EXISTS memberships;
CREATE TABLE memberships(
	org       BINARY(16) NOT NULL,
	user      BINARY(16) NOT NULL,
    PRIMARY KEY (org, user),
    UNIQUE INDEX (user, org),
    FOREIGN KEY (org) REFERENCES accounts (id),
    FOREIGN KEY (user) REFERENCES accounts (id)
);

DROP USER IF EXISTS 'tc_cd'@'%';
CREATE USER 'tc_cd'@'%' IDENTIFIED BY 'T@sk-C3n-T3r'; 
GRANT SELECT ON accounts.accounts TO 'tc_cd'@'%';
GRANT INSERT ON accounts.accounts TO 'tc_cd'@'%';
GRANT UPDATE ON accounts.accounts TO 'tc_cd'@'%';
GRANT DELETE ON accounts.accounts TO 'tc_cd'@'%';
GRANT SELECT ON accounts.users TO 'tc_cd'@'%';
GRANT INSERT ON accounts.users TO 'tc_cd'@'%';
GRANT UPDATE ON accounts.users TO 'tc_cd'@'%';
GRANT DELETE ON accounts.users TO 'tc_cd'@'%';