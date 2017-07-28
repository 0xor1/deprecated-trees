DROP DATABASE IF EXISTS trees;
CREATE DATABASE trees;
USE trees;

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
    totalRemainingTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
    totalLoggedTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
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
    member BINARY(16) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    role TINYINT UNSIGNED NOT NULL, #0 admin, 1 writer, 2 reader
    PRIMARY KEY (account, project, role, member),
    UNIQUE INDEX (account, project, member),
    UNIQUE INDEX (account, member, project)
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
    UNIQUE INDEX(account, project, parent, id),
    UNIQUE INDEX(account, member, project, id),
    UNIQUE INDEX(account, project, member, id)
);

DROP PROCEDURE IF EXISTS registerAccount;
DELIMITER $$
CREATE PROCEDURE registerAccount(_id BINARY(16), _ownerId BINARY(16), _ownerName VARCHAR(50))
BEGIN
	INSERT INTO accounts (id, publicProjectsEnabled) VALUES (_id, false);
    INSERT INTO accountMembers (account, id, name, totalRemainingTime, totalLoggedTime, isActive, role) VALUES (_id, _ownerId, _ownerName, 0, 0, true, 0);
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