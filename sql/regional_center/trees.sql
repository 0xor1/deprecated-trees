DROP DATABASE IF EXISTS trees;
CREATE DATABASE trees;
USE trees;

DROP TABLE IF EXISTS orgs;
CREATE TABLE orgs(
	id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
);

DROP TABLE IF EXISTS orgMembers;
CREATE TABLE orgMembers(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
    name VARCHAR(50) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    isActive BOOL NOT NULL,
    isDeleted BOOL NOT NULL,
    role TINYINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, isActive, role, name),
    UNIQUE INDEX (org, id)
);

DROP TABLE IF EXISTS projects;
CREATE TABLE projects(
	org BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
	member BINARY(16) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    role TINYINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, isActive, role, name),
    UNIQUE INDEX (org, id)
);

DROP TABLE IF EXISTS projectMembers;
CREATE TABLE projectMembers(
	org BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, id),
    UNIQUE INDEX (org, id)
);

DROP TABLE IF EXISTS members_projectAccess;
CREATE TABLE members_projectAccess(
	org BINARY(16) NOT NULL,
    project BINARY(16) NOT NULL,
	member BINARY(16) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    role TINYINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, isActive, role, name),
    UNIQUE INDEX (org, id)
);

DROP TABLE IF EXISTS tasks_noninheritableProperties;
CREATE TABLE tasks_noninheritableProperties(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
    name VARCHAR(250) NOT NULL,
    created DATETIME NOT NULL,
    user BINARY(16) NULL,
    chatCount BIGINT UNSIGNED NOT NULL,
    isAbstractTask BOOL NOT NULL,
    PRIMARY KEY (org, id)
);

DROP TABLE IF EXISTS tasks_timeProperties;
CREATE TABLE tasks_timeProperties(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, id)
);

DROP TABLE IF EXISTS tasks_fileProperties;
CREATE TABLE tasks_fileProperties(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
    fileCount BIGINT UNSIGNED NOT NULL,
    fileSize BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, id)
);

DROP TABLE IF EXISTS abstractTasks_descendantProperties;
CREATE TABLE abstractTasks_descendantProperties(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
	childCount BIGINT UNSIGNED NOT NULL,
	descendantsCount BIGINT UNSIGNED NOT NULL,
	leafCount BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, id)
);

DROP TABLE IF EXISTS abstractTasks_timeProperties;
CREATE TABLE abstractTasks_timeProperties(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
	minimumRemainingTime BIGINT UNSIGNED NOT NULL,
	isParallel BOOL NOT NULL,
    PRIMARY KEY (org, id)
);

DROP TABLE IF EXISTS abstractTasks_fileProperties;
CREATE TABLE abstractTasks_fileProperties(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
	subFileCount BIGINT UNSIGNED NOT NULL,
	subFileSize BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, id)
);

DROP TABLE IF EXISTS abstractTasks_archivedProperties;
CREATE TABLE abstractTasks_archivedProperties(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
	archivedChildCount BIGINT UNSIGNED NOT NULL,
	archivedDescendantCount BIGINT UNSIGNED NOT NULL,
	archivedLeafCount BIGINT UNSIGNED NOT NULL,
	archivedSubFileCount BIGINT UNSIGNED NOT NULL,
	archivedSubFileSize BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, id)
);

DROP TABLE IF EXISTS structure;
CREATE TABLE structure(
	org BINARY(16) NOT NULL,
    project BINARY(16) NOT NULL,
	node BINARY(16) NOT NULL,
	parent BINARY(16) NOT NULL,
    firstChild BINARY(16) NULL,
    nextSibling BINARY(16) NULL,
    PRIMARY KEY (org, node),
    UNIQUE INDEX (org, parent, node),
    UNIQUE INDEX (org, firstChild),
    UNIQUE INDEX (org, nextSibling)
);

DROP PROCEDURE IF EXISTS createAbstractTask;
DELIMITER $$
CREATE PROCEDURE createAbstractTask(_parent BINARY(16), _nextSibling BINARY(16), _org BINARY(16), _id BINARY(16), _name VARCHAR(250), _user BINARY(16), _totalRemainingTime BIGINT UNSIGNED, _totalLoggedTime BIGINT UNSIGNED, _chatCount BIGINT UNSIGNED, _fileCount BIGINT UNSIGNED, _fileSize BIGINT UNSIGNED, _created DATETIME, _isAbstract BOOL, _minimumRemainingTime DATETIME, _isParallel BOOL, _childCount BIGINT UNSIGNED, _descendantCount BIGINT UNSIGNED, _leafCount BIGINT UNSIGNED, _subFileCount BIGINT UNSIGNED, _subFileSize BIGINT UNSIGNED, _archivedChildCount BIGINT UNSIGNED, _archivedDescendantCount BIGINT UNSIGNED, _archivedLeafCount BIGINT UNSIGNED, _archivedSubFileCount BIGINT UNSIGNED, _archivedSubFileSize BIGINT UNSIGNED) 
proc:
BEGIN
	# this proc assumes all the values have already been validated and are appropriate, i.e. the new ids are new and unique
	DECLARE _previous_nextSibling BINARY(16);
    DECLARE _previous_firstChild BINARY(16);
    
	IF _org = _id THEN #creating a new root node for an org
		INSERT INTO tasks(org,	id, name, user, totalRemainingTime, totalLoggedTime, chatCount, fileCount, fileSize, created, isAbstractTask) VALUES (_org,	_id, _name, _user, _totalRemainingTime, _totalLoggedTime, _chatCount, _fileCount, _fileSize, _created, _isAbstractTask);
		INSERT INTO abstractTasks(org, id, minimumRemainingTime, isParallel, childCount, descendantsCount, leafCount, subFileCount, subFileSize, archivedChildCount, archivedDescendantCount, archivedLeafCount, archivedSubFileCount, archivedSubFileSize) VALUES (_org, _id, _minimumRemainingTime, _isParallel, _childCount, _descendantsCount, _leafCount, _subFileCount, _subFileSize, _archivedChildCount, _archivedDescendantCount, _archivedLeafCount, _archivedSubFileCount, _archivedSubFileSize);
		INSERT INTO structure (org, node, parent, firstChild, nextSibling, nodeArchived) VALUES (_org, _id, _parent, NULL, _nexSibling, NULL);
		LEAVE proc;
	END IF;
	
    START TRANSACTION;
	SELECT firstChild INTO _previous_firstChild FROM structure WHERE org = _org AND node = _parent LOCK IN SHARE MODE;
	IF _previous_firstChild IS NULL THEN # this is the first child
	
		UPDATE structure SET firstChild = _id WHERE org = _org AND node = _parent;
		INSERT INTO tasks(org,	id, name, user, totalRemainingTime, totalLoggedTime, chatCount, fileCount, fileSize, created, isAbstractTask) VALUES (_org,	_id, _name, _user, _totalRemainingTime, _totalLoggedTime, _chatCount, _fileCount, _fileSize, _created, _isAbstractTask);
		INSERT INTO abstractTasks(org, id, minimumRemainingTime, isParallel, childCount, descendantsCount, leafCount, subFileCount, subFileSize, archivedChildCount, archivedDescendantCount, archivedLeafCount, archivedSubFileCount, archivedSubFileSize) VALUES (_org, _id, _minimumRemainingTime, _isParallel, _childCount, _descendantsCount, _leafCount, _subFileCount, _subFileSize, _archivedChildCount, _archivedDescendantCount, _archivedLeafCount, _archivedSubFileCount, _archivedSubFileSize);
		INSERT INTO structure (org, node, parent, firstChild, nextSibling, nodeArchived) VALUES (_org, _id, _parent, NULL, _nexSibling, NULL);
		COMMIT;
		LEAVE proc;
		
	END IF;
	
	SELECT node FROM structure WHERE org = _org AND nextSibling = _nextSibling FOR UPDATE;
END;
$$
DELIMITER ;

##this is an internal SP, only to be called by other SPs, it shouldn't start a transaction, it should be running inside the calling SPs transaction.
DROP PROCEDURE IF EXISTS _update_parent_chain_inheritables;
DELIMITER $$
CREATE PROCEDURE _update_parent_chain_inheritables(_node BINARY(16)) 
proc:
BEGIN
	DECLARE _previous_nextSibling BINARY(16);
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