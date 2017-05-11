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
    avatarExt VARCHAR(10) NULL,
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

DROP PROCEDURE IF EXISTS createUser;
DELIMITER $$
CREATE PROCEDURE createUser(_id BINARY(16), _name VARCHAR(50), _created DATETIME, _region CHAR(3), _newRegion CHAR(3), _shard MEDIUMINT, _isUser BOOL, _email VARCHAR(250), _newEmail VARCHAR(250), _activationCode VARCHAR(100), _activated DATETIME, _newEmailConfirmationCode VARCHAR(100), _resetPwdCode VARCHAR(100)) 
BEGIN
	INSERT INTO accounts (id, name, created, region, newRegion, shard, isUser) VALUES (_id, _name, _created, _region, _newRegion, _shard, _isUser);
    INSERT INTO users (id, email, newEmail, activationCode, activated, newEmailConfirmationCode, resetPwdCode) VALUES (_id, _email, _newEmail, _activationCode, _activated, _newEmailConfirmationCode, _resetPwdCode);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS updateUser;
DELIMITER $$
CREATE PROCEDURE updateUser(_id BINARY(16), _name VARCHAR(50), _created DATETIME, _region CHAR(3), _newRegion CHAR(3), _shard MEDIUMINT, _email VARCHAR(250), _newEmail VARCHAR(250), _activationCode VARCHAR(100), _activated DATETIME, _newEmailConfirmationCode VARCHAR(100), _resetPwdCode VARCHAR(100)) 
BEGIN
	UPDATE accounts SET name=_name, created=_created, region=_region, newRegion=_newRegion, shard=_shard WHERE id = _id;
    UPDATE users SET email=_email, newEmail=_newEmail, activationCode=_activationCode, activated=_activated, newEmailConfirmationCode=_newEmailConfirmationCode, resetPwdCode=_resetPwdCode WHERE id = _id;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteUserAndAllAssociatedMemberships;
DELIMITER $$
CREATE PROCEDURE deleteUserAndAllAssociatedMemberships(_id BINARY(16)) 
BEGIN
	DELETE FROM memberships WHERE user = _id;
    DELETE FROM users WHERE id = _id;
    DELETE FROM accounts WHERE id = _id;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS createOrgAndMembership;
DELIMITER $$
CREATE PROCEDURE createOrgAndMembership(_id BINARY(16), _name VARCHAR(50), _created DATETIME, _region CHAR(3), _newRegion CHAR(3), _shard MEDIUMINT, _isUser BOOL, _user BINARY(16)) 
BEGIN
	INSERT INTO accounts (id, name, created, region, newRegion, shard, isUser) VALUES (_id, _name, _created, _region, _newRegion, _shard, _isUser);
    INSERT INTO memberships (org, user) VALUES (_id, _user);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteOrgAndAllAssociatedMemberships;
DELIMITER $$
CREATE PROCEDURE deleteOrgAndAllAssociatedMemberships(_id BINARY(16)) 
BEGIN
	DELETE FROM memberships WHERE org = _id;
    DELETE FROM accounts WHERE id = _id;    
END;
$$
DELIMITER ;

DROP USER IF EXISTS 'tc_cd_accounts'@'%';
CREATE USER 'tc_cd_accounts'@'%' IDENTIFIED BY 'T@sk-C3n-T3r-@cc-0unt5'; 
GRANT SELECT ON accounts.* TO 'tc_cd_accounts'@'%';
GRANT INSERT ON accounts.* TO 'tc_cd_accounts'@'%';
GRANT UPDATE ON accounts.* TO 'tc_cd_accounts'@'%';
GRANT DELETE ON accounts.* TO 'tc_cd_accounts'@'%';
GRANT EXECUTE ON accounts.* TO 'tc_cd_accounts'@'%';