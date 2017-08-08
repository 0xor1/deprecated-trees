DROP DATABASE IF EXISTS trees;
CREATE DATABASE trees;
USE trees;

#BIGINT UNSIGNED duration values are all in units of minutes
#BIGINT UNSIGNED fileSize values are all in units of bytes

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts(
	id BINARY(16) NOT NULL,
    publicProjectsEnabled BOOL NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id)
);

DROP TABLE IF EXISTS accountMembers;
CREATE TABLE accountMembers(
	account BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
    name VARCHAR(50) NOT NULL,
    isActive BOOL NOT NULL DEFAULT TRUE,
    role TINYINT UNSIGNED NOT NULL DEFAULT 2, #0 owner, 1 admin, 2 memberOfAllProjects, 3 memberOfOnlySpecificProjects
    PRIMARY KEY (account, isActive, role, name),
    UNIQUE INDEX (account, isActive, name, role),
    UNIQUE INDEX (account, id)
);

DROP TABLE IF EXISTS accountActivities;
CREATE TABLE accountActivities(
	account BINARY(16) NOT NULL,
    occurredOn BIGINT NOT NULL, #unix millisecs timestamp
    member BINARY(16) NOT NULL,
    item BINARY(16) NOT NULL,
    itemType VARCHAR(100) NOT NULL,
    itemName VARCHAR(250) NOT NULL,
    action VARCHAR(100) NOT NULL,
    newValue VARCHAR(1250) NULL,
    PRIMARY KEY (account, occurredOn, item, member),
    UNIQUE INDEX (account, item, occurredOn, member),
    UNIQUE INDEX (account, member, occurredOn, item)
);

DROP TABLE IF EXISTS projectMembers;
CREATE TABLE projectMembers(
	account BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    name VARCHAR(50) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    role TINYINT UNSIGNED NOT NULL, #0 admin, 1 writer, 2 reader
    PRIMARY KEY (account, project, role, id),
    UNIQUE INDEX (account, project, name, id),
    UNIQUE INDEX (account, project, id),
    UNIQUE INDEX (account, id, project)
);

DROP TABLE IF EXISTS projectActivities;
CREATE TABLE projectActivities(
	account BINARY(16) NOT NULL,
    project BINARY(16) NOT NULL,
    occurredOn BIGINT NOT NULL, #unix millisecs timestamp
    member BINARY(16) NOT NULL,
    item BINARY(16) NOT NULL,
    itemType VARCHAR(100) NOT NULL,
    itemName VARCHAR(250) NOT NULL,
    action VARCHAR(100) NOT NULL,
    newValue VARCHAR(1250) NULL,
    PRIMARY KEY (account, project, occurredOn, item, member),
    UNIQUE INDEX (account, project, item, occurredOn, member),
    UNIQUE INDEX (account, project, member, occurredOn, item),
    UNIQUE INDEX (account, occurredOn, project, item, member)
);

DROP TABLE IF EXISTS projectLocks;
CREATE TABLE projectLocks(
	account BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    PRIMARY KEY(account, id)
);

DROP TABLE IF EXISTS projects;
CREATE TABLE projects(
	account BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    firstChild BINARY(16) NULL,
	name VARCHAR(250) NOT NULL,
	description VARCHAR(1250) NULL,
    createdOn DATETIME NOT NULL,
    archivedOn DateTime NULL,
    startOn DATETIME NULL,
    dueOn DATETIME NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    minimumRemainingTime BIGINT UNSIGNED NOT NULL,
    fileCount BIGINT UNSIGNED NOT NULL,
    fileSize BIGINT UNSIGNED NOT NULL,
    linkedFileCount BIGINT UNSIGNED NOT NULL,
    chatCount BIGINT UNSIGNED NOT NULL,
    isParallel BOOL NOT NULL DEFAULT FALSE,
    isPublic BOOL NOT NULL DEFAULT FALSE,
    PRIMARY KEY (account, id),
    INDEX(account, archivedOn, name, createdOn, id),
    INDEX(account, archivedOn, createdOn, name, id),
    INDEX(account, archivedOn, startOn, name, id),
    INDEX(account, archivedOn, dueOn, name, id),
    INDEX(account, archivedOn, isPublic, name, createdOn, id)
);

DROP TABLE IF EXISTS nodes;
CREATE TABLE nodes(
	account BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    parent BINARY(16) NOT NULL,
    firstChild BINARY(16) NULL,
    nextSibling BINARY(16) NULL,
    isAbstract BOOL NOT NULL,
	name VARCHAR(250) NOT NULL,
	description VARCHAR(1250) NULL,
    createdOn DATETIME NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    minimumRemainingTime BIGINT UNSIGNED NULL,
    linkedFileCount INT UNSIGNED NOT NULL,
    chatCount BIGINT UNSIGNED NOT NULL,
    isParallel BOOL NOT NULL DEFAULT FALSE,
    member BINARY(16) NULL,
    PRIMARY KEY (account, project, id),
    UNIQUE INDEX(account, member, id),
    UNIQUE INDEX(account, project, parent, id),
    UNIQUE INDEX(account, project, member, id)
);

DROP TABLE IF EXISTS timeLogs;
CREATE TABLE timeLogs(
	account BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    node BINARY(16) NOT NULL,
    member BINARY(16) NOT NULL,
    loggedOn DATETIME NOT NULL,
    nodeName VARCHAR(250) NOT NULL,
    duration BIGINT UNSIGNED NOT NULL,
    note VARCHAR(250) NULL,
    PRIMARY KEY(account, project, node, loggedOn, member),
    UNIQUE INDEX(account, project, member, loggedOn, node),
    UNIQUE INDEX(account, member, loggedOn, project, node)
);

DROP PROCEDURE IF EXISTS registerAccount;
DELIMITER $$
CREATE PROCEDURE registerAccount(_id BINARY(16), _ownerId BINARY(16), _ownerName VARCHAR(50))
BEGIN
	INSERT INTO accounts (id, publicProjectsEnabled) VALUES (_id, false);
    INSERT INTO accountMembers (account, id, name, isActive, role) VALUES (_id, _ownerId, _ownerName, true, 0);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS renameMember;
DELIMITER $$
CREATE PROCEDURE renameMember(_account BINARY(16), _member BINARY(16), _newName VARCHAR(50))
BEGIN
	UPDATE accountMembers SET name=_newName WHERE account=_account AND id=_member;
	UPDATE projectMembers SET name=_newName WHERE account=_account AND id=_member;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setMemberInactive;
DELIMITER $$
CREATE PROCEDURE setMemberInactive(_account BINARY(16), _member BINARY(16))
BEGIN
	UPDATE accountMembers SET isActive=false, role=3 WHERE account=_account AND id=_member;
    DELETE FROM projectMembers WHERE account=_account AND id=_member;
    UPDATE nodes SET member=NULL WHERE account=_account AND member=_member;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS createProject;
DELIMITER $$
CREATE PROCEDURE createProject(_accountId BINARY(16), _id BINARY(16), _name VARCHAR(250), _description VARCHAR(1250), _createdOn DATETIME, _archivedOn DATETIME, _startOn DATETIME, _dueOn DATETIME, _totalRemainingTime BIGINT UNSIGNED, _totalLoggedTime BIGINT UNSIGNED, _minimumRemainingTime BIGINT UNSIGNED, _fileCount BIGINT UNSIGNED, _fileSize BIGINT UNSIGNED, _linkedFileCount BIGINT UNSIGNED, _chatCount BIGINT UNSIGNED, _isParallel BOOL, _isPublic BOOL)
BEGIN
	INSERT INTO projectLocks (account, project) VALUES(_accountId, _id);
	INSERT INTO projects (account, id, firstChild, name, description, createdOn, archivedOn, startOn, dueOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, fileCount, fileSize, linkedFileCount, chatCount, isParallel, isPublic) VALUES (_account, _id, _firstChild, _name, _description, _createdOn, _archivedOn, _startOn, _dueOn, _totalRemainingTime, _totalLoggedTime, _minimumRemainingTime, _fileCount, _fileSize, _linkedFileCount, _chatCount, _isParallel, _isPublic);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteAccount;
DELIMITER $$
CREATE PROCEDURE deleteAccount(_id BINARY(16))
BEGIN
	DELETE FROM accounts WHERE id =_id;
	DELETE FROM accountMembers WHERE account =_id;
	DELETE FROM accountActivities WHERE account =_id;
	DELETE FROM projectMembers WHERE account =_id;
	DELETE FROM projectActivities WHERE account =_id;
    DELETE FROM projectLocks WHERE account = _id;
	DELETE FROM projects WHERE account =_id;
	DELETE FROM nodes WHERE account =_id;
	DELETE FROM timeLogs WHERE account =_id;
END;
$$
DELIMITER ;

DROP USER IF EXISTS 'tc_rc_trees'@'%';
CREATE USER 'tc_rc_trees'@'%' IDENTIFIED BY 'T@sk-C3n-T3r-Tr335';
GRANT SELECT ON trees.* TO 'tc_rc_trees'@'%';
GRANT INSERT ON trees.* TO 'tc_rc_trees'@'%';
GRANT UPDATE ON trees.* TO 'tc_rc_trees'@'%';
GRANT DELETE ON trees.* TO 'tc_rc_trees'@'%';
GRANT EXECUTE ON trees.* TO 'tc_rc_trees'@'%';