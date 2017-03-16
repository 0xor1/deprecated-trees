DROP DATABASE IF EXISTS accounts;
CREATE DATABASE accounts;
USE accounts;

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts(
	id BINARY(16) NOT NULL,
    name VARCHAR(50) NOT NULL,
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

DROP USER IF EXISTS 'tc_cd_accounts'@'%';
CREATE USER 'tc_cd_accounts'@'%' IDENTIFIED BY 'T@sk-C3n-T3r'; 
GRANT SELECT ON accounts.* TO 'tc_cd_accounts'@'%';
GRANT INSERT ON accounts.* TO 'tc_cd_accounts'@'%';
GRANT UPDATE ON accounts.* TO 'tc_cd_accounts'@'%';
GRANT DELETE ON accounts.* TO 'tc_cd_accounts'@'%';