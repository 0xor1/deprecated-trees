DROP DATABASE IF EXISTS accounts;
CREATE DATABASE accounts;
USE accounts;

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
    id BINARY(16) NOT NULL,
    name VARCHAR(50) NOT NULL,
    displayName VARCHAR(100) NULL,
    createdOn DATETIME NOT NULL,
    region CHAR(3) NOT NULL,
    newRegion CHAR(3) NULL,
    shard MEDIUMINT NOT NULL DEFAULT -1,
    hasAvatar BOOL NOT NULL DEFAULT FALSE,
    isPersonal BOOL NOT NULL,
    PRIMARY KEY name (name),
    UNIQUE INDEX id (id),
    UNIQUE INDEX displayName_name (displayName, name),
    UNIQUE INDEX isPersonal_name (isPersonal, name),
    UNIQUE INDEX isPersonal_displayName_name (isPersonal, displayName, name)
);

DROP TABLE IF EXISTS personalAccounts;
CREATE TABLE personalAccounts(
	id BINARY(16) NOT NULL,
	email VARCHAR(250) NOT NULL,
	language VARCHAR(50) NOT NULL,
	theme    TINYINT UNSIGNED NOT NULL,
	newEmail VARCHAR(250) NULL,
	activationCode VARCHAR(100) NULL,
	activatedOn DATETIME NULL,
	newEmailConfirmationCode VARCHAR(100) NULL,
	resetPwdCode VARCHAR(100) NULL,
    PRIMARY KEY (id),
    UNIQUE INDEX (email)
);

DROP TABLE IF EXISTS memberships;
CREATE TABLE memberships(
	account   BINARY(16) NOT NULL,
	member      BINARY(16) NOT NULL,
    PRIMARY KEY (account, member),
    UNIQUE INDEX (member, account)
);

DROP PROCEDURE IF EXISTS createPersonalAccount;
DELIMITER $$
CREATE PROCEDURE createPersonalAccount(_id BINARY(16), _name VARCHAR(50), _displayName VARCHAR(100), _createdOn DATETIME, _region CHAR(3), _newRegion CHAR(3), _shard MEDIUMINT, _hasAvatar BOOL, _email VARCHAR(250), _language VARCHAR(50), _theme TINYINT UNSIGNED, _newEmail VARCHAR(250), _activationCode VARCHAR(100), _activatedOn DATETIME, _newEmailConfirmationCode VARCHAR(100), _resetPwdCode VARCHAR(100)) 
BEGIN
	INSERT INTO accounts (id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal) VALUES (_id, _name, _displayName, _createdOn, _region, _newRegion, _shard, _hasAvatar, true);
    INSERT INTO personalAccounts (id, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode) VALUES (_id, _email, _language, _theme, _newEmail, _activationCode, _activatedOn, _newEmailConfirmationCode, _resetPwdCode);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS updatePersonalAccount;
DELIMITER $$
CREATE PROCEDURE updatePersonalAccount(_id BINARY(16), _name VARCHAR(50), _displayName VARCHAR(100), _createdOn DATETIME, _region CHAR(3), _newRegion CHAR(3), _shard MEDIUMINT, _hasAvatar BOOL, _email VARCHAR(250), _language VARCHAR(50), _theme TINYINT UNSIGNED, _newEmail VARCHAR(250), _activationCode VARCHAR(100), _activatedOn DATETIME, _newEmailConfirmationCode VARCHAR(100), _resetPwdCode VARCHAR(100)) 
BEGIN
	UPDATE accounts SET name=_name, displayName=_displayName, createdOn=_createdOn, region=_region, newRegion=_newRegion, shard=_shard, hasAvatar=_hasAvatar WHERE id = _id;
    UPDATE personalAccounts SET email=_email, language=_language, theme=_theme, newEmail=_newEmail, activationCode=_activationCode, activatedOn=_activatedOn, newEmailConfirmationCode=_newEmailConfirmationCode, resetPwdCode=_resetPwdCode WHERE id = _id;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS updateAccountInfo;
DELIMITER $$
CREATE PROCEDURE updateAccountInfo(_id BINARY(16), _name VARCHAR(50), _displayName VARCHAR(100), _createdOn DATETIME, _region CHAR(3), _newRegion CHAR(3), _shard MEDIUMINT, _hasAvatar BOOL, _isPersonal BOOL) 
BEGIN
	UPDATE accounts SET name=_name, displayName=_displayName, createdOn=_createdOn, region=_region, newRegion=_newRegion, shard=_shard, hasAvatar=_hasAvatar WHERE id = _id;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteAccountAndAllAssociatedMemberships;
DELIMITER $$
CREATE PROCEDURE deleteAccountAndAllAssociatedMemberships(_id BINARY(16)) 
BEGIN
	DELETE FROM memberships WHERE account = _id OR member = _id;
    DELETE FROM personalAccounts WHERE id = _id;
    DELETE FROM accounts WHERE id = _id;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS createGroupAccountAndMembership;
DELIMITER $$
CREATE PROCEDURE createGroupAccountAndMembership(_id BINARY(16), _name VARCHAR(50), _displayName VARCHAR(100), _createdOn DATETIME, _region CHAR(3), _newRegion CHAR(3), _shard MEDIUMINT, _hasAvatar BOOL, _member BINARY(16)) 
BEGIN
	INSERT INTO accounts (id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal) VALUES (_id, _name, _displayName, _createdOn, _region, _newRegion, _shard, _hasAvatar, false);
    INSERT INTO memberships (account, member) VALUES (_id, _member);
END;
$$
DELIMITER ;

DROP USER IF EXISTS 't_c_accounts'@'%';
CREATE USER 't_c_accounts'@'%' IDENTIFIED BY 'T@sk-@cc-0unt5';
GRANT SELECT ON accounts.* TO 't_c_accounts'@'%';
GRANT INSERT ON accounts.* TO 't_c_accounts'@'%';
GRANT UPDATE ON accounts.* TO 't_c_accounts'@'%';
GRANT DELETE ON accounts.* TO 't_c_accounts'@'%';
GRANT EXECUTE ON accounts.* TO 't_c_accounts'@'%';