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
    occurredOn BIGINT UNSIGNED NOT NULL, #unix millisecs timestamp
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
    isActive BOOL NOT NULL DEFAULT TRUE,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    role TINYINT UNSIGNED NOT NULL, #0 admin, 1 writer, 2 reader
    PRIMARY KEY (account, project, isActive, role, name),
    UNIQUE INDEX (account, project, isActive, name, role),
    UNIQUE INDEX (account, project, id),
    UNIQUE INDEX (account, id, project)
);

DROP TABLE IF EXISTS projectActivities;
CREATE TABLE projectActivities(
	account BINARY(16) NOT NULL,
    project BINARY(16) NOT NULL,
    occurredOn BIGINT UNSIGNED NOT NULL, #unix millisecs timestamp
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

DROP PROCEDURE IF EXISTS setAccountMemberInactive;
DELIMITER $$
CREATE PROCEDURE setAccountMemberInactive(_account BINARY(16), _member BINARY(16))
BEGIN
	UPDATE accountMembers SET isActive=false, role=3 WHERE account=_account AND id=_member;
    UPDATE projectMembers SET totalRemainingTime=0, totalLoggedTime=0, isActive=false, role=4 WHERE account=_account AND id=_member;
    UPDATE nodes SET member=NULL WHERE account=_account AND member=_member;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteAccount;
DELIMITER $$
CREATE PROCEDURE deleteAccount(_id BINARY(16))
BEGIN
    DELETE FROM projectLocks WHERE account = _id;
	DELETE FROM accounts WHERE id =_id;
	DELETE FROM accountMembers WHERE account =_id;
	DELETE FROM accountActivities WHERE account =_id;
	DELETE FROM projectMembers WHERE account =_id;
	DELETE FROM projectActivities WHERE account =_id;
	DELETE FROM projects WHERE account =_id;
	DELETE FROM nodes WHERE account =_id;
	DELETE FROM timeLogs WHERE account =_id;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS createProject;
DELIMITER $$
CREATE PROCEDURE createProject(_accountId BINARY(16), _id BINARY(16), _name VARCHAR(250), _description VARCHAR(1250), _createdOn DATETIME, _archivedOn DATETIME, _startOn DATETIME, _dueOn DATETIME, _totalRemainingTime BIGINT UNSIGNED, _totalLoggedTime BIGINT UNSIGNED, _minimumRemainingTime BIGINT UNSIGNED, _fileCount BIGINT UNSIGNED, _fileSize BIGINT UNSIGNED, _linkedFileCount BIGINT UNSIGNED, _chatCount BIGINT UNSIGNED, _isParallel BOOL, _isPublic BOOL)
BEGIN
	INSERT INTO projectLocks (account, id) VALUES(_accountId, _id);
	INSERT INTO projects (account, id, firstChild, name, description, createdOn, archivedOn, startOn, dueOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, fileCount, fileSize, linkedFileCount, chatCount, isParallel, isPublic) VALUES (_accountId, _id, NULL, _name, _description, _createdOn, _archivedOn, _startOn, _dueOn, _totalRemainingTime, _totalLoggedTime, _minimumRemainingTime, _fileCount, _fileSize, _linkedFileCount, _chatCount, _isParallel, _isPublic);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setProjectIsParallel;
DELIMITER $$
CREATE PROCEDURE setProjectIsParallel(_accountId BINARY(16), _id BINARY(16), _isParallel BOOL)
BEGIN
	DECLARE projCount TINYINT DEFAULT 0;
    DECLARE childCount INT UNSIGNED DEFAULT 0;
    DECLARE sumChildTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE sumChildMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE maxChildMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    IF _isParallel <> (SELECT isParallel FROM projects WHERE account=_accountId AND id=_id) THEN #make sure we are making a change oterwise, no need to update anything
		START TRANSACTION;
			SELECT COUNT(*) INTO projCount FROM projectLocks WHERE account=_accountId AND id=_id FOR UPDATE; #set project lock to ensure data integrity
			SELECT COUNT(*), SUM(totalRemainingTime), SUM(minimumRemainingTime), MAX(minimumRemainingTime) INTO childCount, sumChildTotalRemainingTime, sumChildMinimumRemainingTime, maxChildMinimumRemainingTime FROM projects WHERE account=_accountId AND id=_id;            
			IF childCount > 0 THEN #settings isParallel and child coutners
				IF _isParallel THEN #setting isParallel to true
					UPDATE projects SET totalRemainingTime=sumChildTotalRemainingTime, minimumRemainingTime=maxChildMinimumRemainingTime, isParallel=_isParallel WHERE account=_accountId AND id=_id;
				ELSE #setting isParallel to false
					UPDATE projects SET totalRemainingTime=sumChildTotalRemainingTime, minimumRemainingTime=sumChildMinimumRemainingTime, isParallel=_isParallel WHERE account=_accountId AND id=_id;
				END IF;
			ELSE #just setting isParallel but not time counters
				UPDATE projects SET isParallel=_isParallel WHERE account=_accountId AND id=_id;
			END IF;
		COMMIT;
    END IF;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteProject;
DELIMITER $$
CREATE PROCEDURE deleteProject(_accountId BINARY(16), _projectId BINARY(16))
BEGIN
    DELETE FROM projectLocks WHERE account=_accountId AND project=_projectId;
	DELETE FROM projectMembers WHERE account=_accountId AND project=_projectId;
	DELETE FROM projectActivities WHERE account=_accountId AND project=_projectId;
	DELETE FROM projects WHERE account=_accountId AND project=_projectId;
	DELETE FROM nodes WHERE account=_accountId AND project=_projectId;
	DELETE FROM timeLogs WHERE account=_accountId AND project=_projectId;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS addProjectMemberOrSetActive;
DELIMITER $$ #TRIES to add a user to a project, sets true if they were added, false otherwise. It may be false if they are trying to add a user who is not a member of the account, or add a member who is already an active member of this project.
CREATE PROCEDURE addProjectMemberOrSetActive(_accountId BINARY(16), _projectId BINARY(16), _id BINARY(16), _role TINYINT UNSIGNED)
BEGIN
	DECLARE projMemberCount TINYINT DEFAULT 0;
	DECLARE projMemberIsActive BOOL DEFAULT false;
	DECLARE accMemberName VARCHAR(50) DEFAULT '';
    SELECT COUNT(*), isActive INTO projMemberCount, projMemberIsActive FROM projectMembers WHERE account = _accountId AND project = _projectId AND id = _id;
    IF projMemberCount = 1 AND projMemberIsActive = false THEN #setting previous member back to active, still need to check if they are an active account member
		IF (SELECT COUNT(*) FROM accountMembers WHERE account = _accountId AND id = _id AND isActive = true) THEN #if active account member then add them to the project
			UPDATE projectMembers SET role = _role, isActive = true WHERE account = _accountId AND project = _projectId AND id = _id;
            SELECT true;
        ELSE #they are a disabled account member and so can not be added to the project
			SELECT false;
		END IF;
	ELSEIF projMemberCount = 1 AND projMemberIsActive = true THEN #they are already an active member of this project
		SELECT false;
    ELSEIF projMemberCount = 0 THEN #adding new project member, need to check if they are active account member
		START TRANSACTION;
			SELECT name INTO accMemberName FROM accountMembers WHERE account = _accountId AND id = _id AND isActive = true LOCK IN SHARE MODE;
			IF accMemberName IS NOT NULL AND accMemberName <> '' THEN #if active account member then add them to the project
				INSERT INTO projectMembers (account, project, id, name, isActive, totalRemainingTime, totalLoggedTime, role) VALUES (_accountId, _projectId, _id, accMemberName, true, 0, 0, _role);
				SELECT true;
			ELSE #they are a not an active account member so return false
				SELECT false;
			END IF;
        COMMIT;
    END IF;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setProjectMemberInactive;
DELIMITER $$
CREATE PROCEDURE setProjectMemberInactive(_accountId BINARY(16), _projectId BINARY(16), _id BINARY(16))
BEGIN
	DECLARE projMemberCount TINYINT DEFAULT 0;	
	START TRANSACTION;
		SELECT COUNT(*) INTO projMemberCount FROM projectMembers WHERE account = _accountId AND project = _projectId AND id = _id AND isActive = true FOR UPDATE;
		IF projMemberCount = 1 THEN
			UPDATE nodes SET member = NULL WHERE account = _accountId AND project = _projectId AND member = _id;
            UPDATE projectMembers SET totalRemainingTime = 0, totalLoggedTime = 0 WHERE account = _account AND project = _projectId AND member = _id;
            SELECT true;
		ELSE
			SELECT false;
        END IF;
    COMMIT;
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