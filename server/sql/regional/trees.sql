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
    displayName VARCHAR(100) NULL,
    isActive BOOL NOT NULL DEFAULT TRUE,
    role TINYINT UNSIGNED NOT NULL DEFAULT 2, #0 owner, 1 admin, 2 memberOfAllProjects, 3 memberOfOnlySpecificProjects
    PRIMARY KEY (account, isActive, role, name),
    UNIQUE INDEX (account, isActive, role, displayName, name),
    UNIQUE INDEX (account, isActive, name, role),
    UNIQUE INDEX (account, isActive, displayName, role, name),
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
    displayName VARCHAR(100) NULL,
    isActive BOOL NOT NULL DEFAULT TRUE,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    role TINYINT UNSIGNED NOT NULL, #0 admin, 1 writer, 2 reader
    PRIMARY KEY (account, project, isActive, role, name),
    UNIQUE INDEX (account, project, isActive, role, displayName, name),
    UNIQUE INDEX (account, project, isActive, name, role),
    UNIQUE INDEX (account, project, isActive, displayName, role, name),
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
    childCount BIGINT UNSIGNED NOT NULL,
    descendantCount BIGINT UNSIGNED NOT NULL,
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
    minimumRemainingTime BIGINT UNSIGNED NOT NULL,
    linkedFileCount INT UNSIGNED NOT NULL,
    chatCount BIGINT UNSIGNED NOT NULL,
    childCount BIGINT UNSIGNED NULL,
    descendantCount BIGINT UNSIGNED NULL,
    isParallel BOOL NULL DEFAULT FALSE,
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
CREATE PROCEDURE registerAccount(_id BINARY(16), _ownerId BINARY(16), _ownerName VARCHAR(50), _ownerDisplayName VARCHAR(100))
BEGIN
	INSERT INTO accounts (id, publicProjectsEnabled) VALUES (_id, false);
    INSERT INTO accountMembers (account, id, name, displayName, isActive, role) VALUES (_id, _ownerId, _ownerName, _ownerDisplayName, true, 0);
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

DROP PROCEDURE IF EXISTS setMemberDisplayName;
DELIMITER $$
CREATE PROCEDURE setMemberDisplayName(_account BINARY(16), _member BINARY(16), _newDisplayName VARCHAR(100))
BEGIN
	UPDATE accountMembers SET displayName=_newDisplayName WHERE account=_account AND id=_member;
	UPDATE projectMembers SET displayName=_newDisplayName WHERE account=_account AND id=_member;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setAccountMemberInactive;
DELIMITER $$
CREATE PROCEDURE setAccountMemberInactive(_account BINARY(16), _member BINARY(16))
BEGIN
	UPDATE accountMembers SET isActive=false, role=3 WHERE account=_account AND id=_member;
    UPDATE projectMembers SET isActive=false, role=2 WHERE account=_account AND id=_member;
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
CREATE PROCEDURE createProject(_accountId BINARY(16), _id BINARY(16), _name VARCHAR(250), _description VARCHAR(1250), _createdOn DATETIME, _archivedOn DATETIME, _startOn DATETIME, _dueOn DATETIME, _totalRemainingTime BIGINT UNSIGNED, _totalLoggedTime BIGINT UNSIGNED, _minimumRemainingTime BIGINT UNSIGNED, _fileCount BIGINT UNSIGNED, _fileSize BIGINT UNSIGNED, _linkedFileCount BIGINT UNSIGNED, _chatCount BIGINT UNSIGNED, _childCount BIGINT UNSIGNED, _descendantCount BIGINT UNSIGNED, _isParallel BOOL, _isPublic BOOL)
BEGIN
	INSERT INTO projectLocks (account, id) VALUES(_accountId, _id);
	INSERT INTO projects (account, id, firstChild, name, description, createdOn, archivedOn, startOn, dueOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, fileCount, fileSize, linkedFileCount, chatCount, childCount, descendantCount, isParallel, isPublic) VALUES (_accountId, _id, NULL, _name, _description, _createdOn, _archivedOn, _startOn, _dueOn, _totalRemainingTime, _totalLoggedTime, _minimumRemainingTime, _fileCount, _fileSize, _linkedFileCount, _chatCount, _childCount, _descendantCount, _isParallel, _isPublic);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setProjectIsParallel;
DELIMITER $$
CREATE PROCEDURE setProjectIsParallel(_accountId BINARY(16), _id BINARY(16), _isParallel BOOL)
BEGIN
	DECLARE projCount TINYINT DEFAULT 0;
    DECLARE childCount INT UNSIGNED DEFAULT 0;
    DECLARE sumChildMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE maxChildMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    IF _isParallel <> (SELECT isParallel FROM projects WHERE account=_accountId AND id=_id) THEN #make sure we are making a change oterwise, no need to update anything
		START TRANSACTION;
			SELECT COUNT(*) INTO projCount FROM projectLocks WHERE account=_accountId AND id=_id FOR UPDATE; #set project lock to ensure data integrity
			SELECT COUNT(*), SUM(minimumRemainingTime), MAX(minimumRemainingTime) INTO childCount, sumChildMinimumRemainingTime, maxChildMinimumRemainingTime FROM projects WHERE account=_accountId AND id=_id;            
			IF childCount > 0 THEN #settings isParallel and child coutners
				IF _isParallel THEN #setting isParallel to true
					UPDATE projects SET minimumRemainingTime=maxChildMinimumRemainingTime, isParallel=_isParallel WHERE account=_accountId AND id=_id;
				ELSE #setting isParallel to false
					UPDATE projects SET minimumRemainingTime=sumChildMinimumRemainingTime, isParallel=_isParallel WHERE account=_accountId AND id=_id;
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
    DELETE FROM projectLocks WHERE account=_accountId AND id=_projectId;
	DELETE FROM projectMembers WHERE account=_accountId AND project=_projectId;
	DELETE FROM projectActivities WHERE account=_accountId AND project=_projectId;
	DELETE FROM projects WHERE account=_accountId AND id=_projectId;
	DELETE FROM nodes WHERE account=_accountId AND project=_projectId;
	DELETE FROM timeLogs WHERE account=_accountId AND project=_projectId;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS addProjectMemberOrSetActive;
DELIMITER $$
CREATE PROCEDURE addProjectMemberOrSetActive(_accountId BINARY(16), _projectId BINARY(16), _id BINARY(16), _role TINYINT UNSIGNED)
BEGIN
	DECLARE projMemberCount TINYINT DEFAULT 0;
	DECLARE projMemberIsActive BOOL DEFAULT false;
	DECLARE accMemberName VARCHAR(50) DEFAULT '';
	DECLARE accMemberDisplayName VARCHAR(100) DEFAULT NULL;
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
			SELECT name, displayName INTO accMemberName, accMemberDisplayName FROM accountMembers WHERE account = _accountId AND id = _id AND isActive = true LOCK IN SHARE MODE;
			IF accMemberName IS NOT NULL AND accMemberName <> '' THEN #if active account member then add them to the project
				INSERT INTO projectMembers (account, project, id, name, displayName, isActive, totalRemainingTime, totalLoggedTime, role) VALUES (_accountId, _projectId, _id, accMemberName, accMemberDisplayName, true, 0, 0, _role);
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
            UPDATE projectMembers SET totalRemainingTime = 0 WHERE account = _accountId AND project = _projectId AND id = _id;
            SELECT true;
		ELSE
			SELECT false;
        END IF;
    COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS createNode;
DELIMITER $$
CREATE PROCEDURE createNode(_accountId BINARY(16), _projectId BINARY(16), _parentId BINARY(16), _previousSiblingId BINARY(16), _nodeId BINARY(16), _isAbstract BOOLEAN, _name VARCHAR(250), _description VARCHAR(1250), _createdOn DATETIME, _totalRemainingTime BIGINT UNSIGNED, _totalLoggedTime BIGINT UNSIGNED, _minimumRemainingTime BIGINT UNSIGNED, _linkedFileCount INT UNSIGNED, _chatCount BIGINT UNSIGNED, _childCount BIGINT UNSIGNED, _descendantCount BIGINT UNSIGNED, _isParallel BOOLEAN, _memberId BINARY(16))
CONTAINS SQL `createNode`:
BEGIN
	DECLARE originalParentId BINARY(16) DEFAULT _parentId;
	DECLARE projectExists BOOLEAN DEFAULT FALSE;
	DECLARE parentExists BOOLEAN DEFAULT FALSE;
	DECLARE previousSiblingExists BOOLEAN DEFAULT FALSE;
	DECLARE nextSiblingIdToUse BINARY(16) DEFAULT NULL;
    DECLARE memberExistsAndIsActive BOOLEAN DEFAULT FALSE;
    #necessary values for traveling up the tree
    DECLARE currentIsParallel BOOLEAN DEFAULT FALSE;
    DECLARE currentMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
	#assume parameters are already validated against themselves in the business logic
    START TRANSACTION;
    #validate parameters against the database, i.e. make sure parent/sibiling/member exist and are active and have sufficient privelages 
	#lock the project and lock the project member (member needs "for update" lock if remaining time is greater than zero, otherwise only "lock in share mode" is needed)
	SELECT COUNT(*) = 1 INTO projectExists FROM projectLocks WHERE account = _accountId AND id = _projectId FOR UPDATE;
    IF _parentId = _projectId THEN
		SET parentExists = projectExists;
    ELSE
		SELECT COUNT(*) = 1 INTO parentExists FROM nodes WHERE account = _accountId AND project = _projectId AND id = _parentId AND isAbstract = TRUE;	
    END IF;
    IF _previousSiblingId IS NULL THEN
		IF _parentId = _projectId THEN
			SELECT firstChild INTO nextSiblingIdToUse FROM projects WHERE account = _accountId AND id = _projectId;
        ELSE
			SELECT firstChild INTO nextSiblingIdToUse FROM nodes WHERE account = _accountId AND project = _projectId AND id = _parentId;        
        END IF;
		SET previousSiblingExists = TRUE;
	ELSE
        SELECT COUNT(*) = 1, nextSibling INTO previousSiblingExists, nextSiblingIdToUse FROM nodes WHERE account = _accountId AND project = _projectId AND id = _previousSiblingId;
    END IF;
    IF _memberId IS NOT NULL AND _totalRemainingTime <> 0 THEN
		SELECT COUNT(*) = 1 INTO memberExistsAndIsActive FROM projectMembers WHERE account = _accountId AND project = _projectId AND id = _memberId AND isActive = TRUE AND role < 2 FOR UPDATE; #less than 2 means 0->projectAdmin or 1->projectWriter
    ELSEIF _memberId IS NOT NULL AND _totalRemainingTime = 0 THEN
		SELECT COUNT(*) = 1 INTO memberExistsAndIsActive FROM projectMembers WHERE account = _accountId AND project = _projectId AND id = _memberId AND isActive = TRUE AND role < 2 LOCK IN SHARE MODE;
    ELSE
		SET memberExistsAndIsActive = TRUE;
    END IF;
    IF NOT(projectExists AND parentExists AND previousSiblingExists AND memberExistsAndIsActive) THEN
		#something is wrong, exit
		SELECT FALSE;
	ELSE
		#write the node row
        INSERT INTO nodes (account,	project, id, parent, firstChild, nextSibling, isAbstract, name,	description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member) VALUES (_accountId,	_projectId, _nodeId, _parentId, NULL, nextSiblingIdToUse, _isAbstract, _name, _description, _createdOn, _totalRemainingTime, _totalLoggedTime, _minimumRemainingTime, _linkedFileCount, _chatCount, _childCount, _descendantCount, _isParallel, _memberId);
        #update siblings and parent firstChild value if required
        IF _previousSiblingId IS NULL THEN #update parents firstChild
			IF _parentId = _projectId THEN
				UPDATE projects SET firstChild = _nodeId WHERE account = _accountId AND id = _projectId;
			ELSE
				UPDATE nodes SET firstChild = _nodeId WHERE account = _accountId AND project = _projectId AND id = _parentId;
			END IF;
        ELSE
			UPDATE nodes SET nextSibling = nodeId WHERE account = _accountId AND project = _projectId AND id = _previousSiblingId;
        END IF;
        #update member if needed
        IF _memberId IS NOT NULL AND _totalRemainingTime <> 0 THEN
			UPDATE projectMembers SET totalRemainingTime = totalRemainingTime + _totalRemainingTime WHERE account = _accountId AND project = _projectId AND id = _memberId;
        END IF;
        WHILE _parentId IS NOT NULL DO
			IF _parentId = _projectId THEN #UPDATE project node
				IF _totalRemainingTime = 0 THEN #UPDATE childCount and descendantCount
					IF _parentId = originalParentId THEN #UPDATE childCount and descendantCount
						UPDATE projects SET childCount = childCount + 1, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                    ELSE #UPDATE descendantCount
						UPDATE projects SET descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                    END IF;
                ELSE #UPDATE time values AND childCount and descendantCount
					SELECT isParallel, minimumRemainingTime INTO currentIsParallel, currentMinimumRemainingTime FROM projects WHERE account = _accountId AND id = _projectId;
					IF _parentId = originalParentId THEN #UPDATE time values AND childCount and descendantCount
						IF currentIsParallel THEN
							IF _minimumRemainingTime > currentMinimumRemainingTime THEN #UPDATE total and minimum times and childCount and descendantCount
								UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, minimumRemainingTime = _minimumRemainingTime, childCount = childCount + 1, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                            ELSE #UPDATE total times and childCount and descendantCount 
								UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, childCount = childCount + 1, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                            END IF;
                        ELSE
							UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, minimumRemainingTime = minimumRemainingTime + _minimumRemainingTime, childCount = childCount + 1, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                        END IF;
                    ELSE #UPDATE time values AND descendantCount
						IF currentIsParallel THEN
							IF _minimumRemainingTime > currentMinimumRemainingTime THEN #UPDATE total and minimum times and childCount and descendantCount
								UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, minimumRemainingTime = _minimumRemainingTime, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                            ELSE #UPDATE total times and childCount and descendantCount 
								UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                            END IF;
                        ELSE
							UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, minimumRemainingTime = minimumRemainingTime + _minimumRemainingTime, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                        END IF;
                    END IF;
				END IF;
                SET _parentId = NULL;
            ELSE #UPDATE abstract node
				SELECT isParallel, minimumRemainingTime, parent INTO currentIsParallel, currentMinimumRemainingTime, nextSiblingIdToUse FROM nodes WHERE account = _accountId AND project = _projectId AND id = _parentId; #nextSiblingIdToUse var is pulling double duty here, it has already served its purpose and is now being used as a temporary holding value for the next _parentId value
				IF _totalRemainingTime = 0 THEN #UPDATE childCount and descendantCount
					IF _parentId = originalParentId THEN #UPDATE childCount and descendantCount
						UPDATE nodes SET childCount = childCount + 1, descendantCount = descendantCount + 1 WHERE account = _accountId AND project = _projectId AND id = _parentId;
                    ELSE #UPDATE descendantCount
						UPDATE nodes SET descendantCount = descendantCount + 1 WHERE account = _accountId AND project = _projectId AND id = _parentId;
                    END IF;
					SET _minimumRemainingTime = currentMinimumRemainingTime;
                ELSE #UPDATE time values AND childCount and descendantCount
					SELECT isParallel, minimumRemainingTime INTO currentIsParallel, currentMinimumRemainingTime FROM nodes WHERE account = _accountId AND project = _projectId AND id = _parentId;
					IF _parentId = originalParentId THEN #UPDATE time values AND childCount and descendantCount
						IF currentIsParallel THEN
							IF _minimumRemainingTime > currentMinimumRemainingTime THEN #UPDATE total and minimum times and childCount and descendantCount
								UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, minimumRemainingTime = _minimumRemainingTime, childCount = childCount + 1, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                            ELSE #UPDATE total times and childCount and descendantCount 
								UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, childCount = childCount + 1, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
								SET _minimumRemainingTime = currentMinimumRemainingTime;
                            END IF;
                        ELSE
							UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, minimumRemainingTime = minimumRemainingTime + _minimumRemainingTime, childCount = childCount + 1, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
							SET _minimumRemainingTime = currentMinimumRemainingTime + _minimumRemainingTime;
                        END IF;
                    ELSE #UPDATE time values AND descendantCount
						IF currentIsParallel THEN
							IF _minimumRemainingTime > currentMinimumRemainingTime THEN #UPDATE total and minimum times and childCount and descendantCount
								UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, minimumRemainingTime = _minimumRemainingTime, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
                            ELSE #UPDATE total times and childCount and descendantCount 
								UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
								SET _minimumRemainingTime = currentMinimumRemainingTime;
                            END IF;
                        ELSE
							UPDATE projects SET totalRemainingTime = totalRemainingTime + _totalRemainingTime, minimumRemainingTime = minimumRemainingTime + _minimumRemainingTime, descendantCount = descendantCount + 1 WHERE account = _accountId AND id = _projectId;
							SET _minimumRemainingTime = currentMinimumRemainingTime + _minimumRemainingTime;
                        END IF;
                    END IF;
				END IF;
                SET _parentId = nextSiblingIdToUse;
            END IF;
        END WHILE;
		SELECT TRUE;
    END IF;
    
    COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setNodeIsParallel
DELIMITER $$
CREATE PROCEDURE setNodeIsParallel(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _isParallel BOOLEAN)
BEGIN
	#TODO
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setNodeMember
DELIMITER $$
CREATE PROCEDURE setNodeMember(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _memberId BINARY(16))
BEGIN
	#TODO
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setTimeRemaining
DELIMITER $$
CREATE PROCEDURE setTimeRemaining(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _timeRemaining BIGINT UNSIGNED)
BEGIN
	#TODO
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS logTimeAndSetTimeRemaining
DELIMITER $$
CREATE PROCEDURE logTimeAndSetTimeRemaining(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _myId BINARY(16), _timeRemaining BIGINT UNSIGNED, _note VARCHAR(250))
BEGIN
	#TODO
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS moveNode
DELIMITER $$
CREATE PROCEDURE moveNode(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _parentId BINARY(16), _nextSibling BINARY(16))
BEGIN
	#TODO
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteNode
DELIMITER $$
CREATE PROCEDURE deleteNode(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16))
BEGIN
	#TODO
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS getNodes
DELIMITER $$
CREATE PROCEDURE getNodes(_accountId BINARY(16), _projectId BINARY(16), _parentId BINARY(16), _fromSiblingId BINARY(16), _limit INT)
BEGIN
	#TODO
END;
$$
DELIMITER ;

DROP USER IF EXISTS 't_r_trees'@'%';
CREATE USER 't_r_trees'@'%' IDENTIFIED BY 'T@sk-Tr335';
GRANT SELECT ON trees.* TO 't_r_trees'@'%';
GRANT INSERT ON trees.* TO 't_r_trees'@'%';
GRANT UPDATE ON trees.* TO 't_r_trees'@'%';
GRANT DELETE ON trees.* TO 't_r_trees'@'%';
GRANT EXECUTE ON trees.* TO 't_r_trees'@'%';