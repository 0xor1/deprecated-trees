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
  occurredOn DATETIME(6) NOT NULL,
  member BINARY(16) NOT NULL,
  item BINARY(16) NOT NULL,
  itemType VARCHAR(100) NOT NULL,
  action VARCHAR(100) NOT NULL,
  itemName VARCHAR(250) NULL,
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
  occurredOn DATETIME(6) NOT NULL,
  member BINARY(16) NOT NULL,
  item BINARY(16) NOT NULL,
  itemType VARCHAR(100) NOT NULL,
  action VARCHAR(100) NOT NULL,
  itemName VARCHAR(250) NULL,
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
  isArchived BOOLEAN NOT NULL,
	name VARCHAR(250) NOT NULL,
  createdOn DATETIME NOT NULL,
  startOn DATETIME NULL,
  dueOn DATETIME NULL,
  fileCount BIGINT UNSIGNED NOT NULL,
  fileSize BIGINT UNSIGNED NOT NULL,
  isPublic BOOL NOT NULL DEFAULT FALSE,
  PRIMARY KEY (account, id),
  INDEX(account, isArchived, name, createdOn, id),
  INDEX(account, isArchived, createdOn, name, id),
  INDEX(account, isArchived, startOn, name, id),
  INDEX(account, isArchived, dueOn, name, id),
  INDEX(account, isArchived, isPublic, name, createdOn, id)
);

DROP TABLE IF EXISTS nodes;
CREATE TABLE nodes(
	account BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
  id BINARY(16) NOT NULL,
  parent BINARY(16) NULL,
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
  childCount BIGINT UNSIGNED NOT NULL,
  descendantCount BIGINT UNSIGNED NOT NULL,
  isParallel BOOL NOT NULL DEFAULT FALSE,
  member BINARY(16) NULL,
  PRIMARY KEY (account, project, id),
  UNIQUE INDEX(account, member, id),
  UNIQUE INDEX(account, project, parent, id),
  UNIQUE INDEX(account, project, nextSibling, id),
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
CREATE PROCEDURE registerAccount(_accountId BINARY(16), _myId BINARY(16), _myName VARCHAR(50), _myDisplayName VARCHAR(100))
BEGIN
	INSERT INTO accounts (id, publicProjectsEnabled) VALUES (_accountId, false);
  INSERT INTO accountMembers (account, id, name, displayName, isActive, role) VALUES (_accountId, _myId, _myName, _myDisplayName, true, 0);
  INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _accountId, 'account', 'created', NULL, NULL);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setMemberName;
DELIMITER $$
CREATE PROCEDURE setMemberName(_accountId BINARY(16), _memberId BINARY(16), _newName VARCHAR(50))
BEGIN
	UPDATE accountMembers SET name=_newName WHERE account=_accountId AND id=_memberId;
	UPDATE projectMembers SET name=_newName WHERE account=_accountId AND id=_memberId;
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
  UPDATE accountMembers SET isActive=FALSE, role=3 WHERE account=_account AND id=_member;
  UPDATE projectMembers SET isActive=FALSE, totalRemainingTime=0, role=2 WHERE account=_account AND id=_member;
  UPDATE nodes SET member=NULL WHERE account=_account AND member=_member;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS updateMembersAndSetActive;
DELIMITER $$
CREATE PROCEDURE updateMembersAndSetActive(_account BINARY(16), _member BINARY(16), _memberName VARCHAR(50), _displayName VARCHAR(100), _role TINYINT UNSIGNED)
  BEGIN
    UPDATE accountMembers SET isActive=TRUE, role=_role, name=_memberName, displayName=_displayName WHERE account=_account AND id=_member;
    UPDATE projectMembers SET name=_memberName, displayName=_displayName WHERE account=_account AND id=_member;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteAccount;
DELIMITER $$
CREATE PROCEDURE deleteAccount(_id BINARY(16))
  BEGIN
    DELETE FROM projectLocks WHERE account=_id;
    DELETE FROM accounts WHERE id=_id;
    DELETE FROM accountMembers WHERE account=_id;
    DELETE FROM accountActivities WHERE account=_id;
    DELETE FROM projectMembers WHERE account=_id;
    DELETE FROM projectActivities WHERE account=_id;
    DELETE FROM projects WHERE account=_id;
    DELETE FROM nodes WHERE account=_id;
    DELETE FROM timeLogs WHERE account=_id;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setPublicProjectsEnabled;
DELIMITER $$
CREATE PROCEDURE setPublicProjectsEnabled(_accountId BINARY(16), _myId BINARY(16), _enabled BOOLEAN)
  BEGIN
    UPDATE accounts SET publicProjectsEnabled=_enabled WHERE id=_accountId;
    IF NOT _enabled THEN
      UPDATE projects SET isPublic=FALSE WHERE account=_accountId;
    END IF;
    IF _enabled THEN
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _accountId, 'account', 'setPublicProjectsEnabled', NULL, 'true');
    ELSE
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _accountId, 'account', 'setPublicProjectsEnabled', NULL, 'false');
    END IF;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setAccountMemberRole;
DELIMITER $$
CREATE PROCEDURE setAccountMemberRole(_accountId BINARY(16), _myId BINARY(16), _memberId BINARY(16), _role TINYINT UNSIGNED)
  BEGIN
    DECLARE memberExists BOOLEAN DEFAULT FALSE;
    SELECT COUNT(*)=1 INTO memberExists  FROM accountMembers WHERE account=_accountId AND id=_memberId AND isActive=TRUE FOR UPDATE;
    START TRANSACTION;
    IF memberExists THEN
      UPDATE accountMembers SET role=_role WHERE account=_accountId AND id=_memberId AND isActive=TRUE;
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _memberId, 'member', 'setAccountRole', NULL, CAST(_role as char character set utf8));
    END IF;
    COMMIT;
    SELECT memberExists;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS createProject;
DELIMITER $$
CREATE PROCEDURE createProject(_accountId BINARY(16), _id BINARY(16), _myId BINARY(16), _name VARCHAR(250), _description VARCHAR(1250), _createdOn DATETIME, _startOn DATETIME, _dueOn DATETIME, _isParallel BOOL, _isPublic BOOL)
  BEGIN
    INSERT INTO projectLocks (account, id) VALUES(_accountId, _id);
    INSERT INTO projects (account, id, isArchived, name, createdOn, startOn, dueOn, fileCount, fileSize, isPublic) VALUES (_accountId, _id, FALSE, _name, _createdOn, _startOn, _dueOn, 0, 0, _isPublic);
    INSERT INTO nodes (account, project, id, parent, firstChild, nextSibling, isAbstract, name, description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member) VALUES (_accountId, _id, _id, NULL, NULL, NULL, TRUE, _name, _description, _createdOn, 0, 0, 0, 0, 0, 0, 0, _isParallel, NULL);
    INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _id, 'project', 'created', _name, NULL);
    INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _id, UTC_TIMESTAMP(6), _myId, _id, 'project', 'created', NULL, NULL);
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setProjectIsPublic;
DELIMITER $$
CREATE PROCEDURE setProjectIsPublic(_accountId BINARY(16), _projectId BINARY(16), _myId BINARY(16), _isPublic BOOL)
  BEGIN
    DECLARE projName VARCHAR(250);
    SELECT name INTO projName FROM projects WHERE account=_accountId AND id=_projectId;
    UPDATE projects SET isPublic=_isPublic WHERE account=_accountId AND id=_projectId;
    IF _isPublic THEN
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsPublic', projName, 'true');
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsPublic', NULL, 'true');
    ELSE
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsPublic', projName, 'false');
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsPublic', NULL, 'false');
    END IF;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setProjectIsArchived;
DELIMITER $$
CREATE PROCEDURE setProjectIsArchived(_accountId BINARY(16), _projectId BINARY(16), _myId BINARY(16), _isArchived BOOL)
  BEGIN
    DECLARE projName VARCHAR(250);
    SELECT name INTO projName FROM projects WHERE account=_accountId AND id=_projectId;
    UPDATE projects SET isArchived=_isArchived WHERE account=_accountId AND id=_projectId;
    IF _isArchived THEN
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsArchived', projName, 'true');
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsArchived', NULL, 'true');
    ELSE
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsArchived', projName, 'false');
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsArchived', NULL, 'false');
    END IF;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteProject;
DELIMITER $$
CREATE PROCEDURE deleteProject(_accountId BINARY(16), _projectId BINARY(16), _myId BINARY(16))
BEGIN
  DECLARE projName VARCHAR(250);
  SELECT name INTO projName FROM projects WHERE account=_accountId AND id=_projectId;
  DELETE FROM projectLocks WHERE account=_accountId AND id=_projectId;
	DELETE FROM projectMembers WHERE account=_accountId AND project=_projectId;
	DELETE FROM projectActivities WHERE account=_accountId AND project=_projectId;
	DELETE FROM projects WHERE account=_accountId AND id=_projectId;
	DELETE FROM nodes WHERE account=_accountId AND project=_projectId;
	DELETE FROM timeLogs WHERE account=_accountId AND project=_projectId;
  INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'deleted', projName, NULL);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS addProjectMemberOrSetActive;
DELIMITER $$
CREATE PROCEDURE addProjectMemberOrSetActive(_accountId BINARY(16), _projectId BINARY(16), _myId BINARY(16), _id BINARY(16), _role TINYINT UNSIGNED)
BEGIN
	DECLARE projMemberCount TINYINT DEFAULT 0;
	DECLARE projMemberIsActive BOOL DEFAULT false;
	DECLARE accMemberName VARCHAR(50) DEFAULT '';
	DECLARE accMemberDisplayName VARCHAR(100) DEFAULT NULL;
  SELECT COUNT(*), isActive INTO projMemberCount, projMemberIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_id;
  IF projMemberCount=1 AND projMemberIsActive=false THEN #setting previous member back to active, still need to check if they are an active account member
    IF (SELECT COUNT(*) FROM accountMembers WHERE account=_accountId AND id=_id AND isActive=true) THEN #if active account member then add them to the project
      UPDATE projectMembers SET role=_role, isActive=true WHERE account=_accountId AND project=_projectId AND id=_id;
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _id, 'member', 'added', NULL, NULL);
      SELECT true;
    ELSE #they are a disabled account member and so can not be added to the project
      SELECT false;
    END IF;
	ELSEIF projMemberCount=1 AND projMemberIsActive=true THEN #they are already an active member of this project
		SELECT false;
  ELSEIF projMemberCount=0 THEN #adding new project member, need to check if they are active account member
		START TRANSACTION;
			SELECT name, displayName INTO accMemberName, accMemberDisplayName FROM accountMembers WHERE account=_accountId AND id=_id AND isActive=true LOCK IN SHARE MODE;
			IF accMemberName IS NOT NULL AND accMemberName <> '' THEN #if active account member then add them to the project
				INSERT INTO projectMembers (account, project, id, name, displayName, isActive, totalRemainingTime, totalLoggedTime, role) VALUES (_accountId, _projectId, _id, accMemberName, accMemberDisplayName, true, 0, 0, _role);
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _id, 'member', 'added', NULL, NULL);
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
CREATE PROCEDURE setProjectMemberInactive(_accountId BINARY(16), _projectId BINARY(16), _myId BINARY(16), _id BINARY(16))
BEGIN
  DECLARE projExists BOOLEAN DEFAULT FALSE;
  DECLARE projMemberExists BOOLEAN DEFAULT FALSE;

  START TRANSACTION;
    SELECT COUNT(*)=1 INTO projExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
		SELECT COUNT(*)=1 INTO projMemberExists FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_id AND isActive=true FOR UPDATE;
		IF projMemberExists THEN
			UPDATE nodes SET member=NULL WHERE account=_accountId AND project=_projectId AND member=_id;
      UPDATE projectMembers SET totalRemainingTime=0 WHERE account=_accountId AND project=_projectId AND id=_id;
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _id, 'member', 'removed', NULL, NULL);
      SELECT true;
		ELSE
			SELECT false;
    END IF;
    COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setProjectMemberRole;
DELIMITER $$
CREATE PROCEDURE setProjectMemberRole(_accountId BINARY(16), _projectId BINARY(16), _myId BINARY(16), _memberId BINARY(16), _role TINYINT UNSIGNED)
  BEGIN
    DECLARE memberExists BOOLEAN DEFAULT FALSE;
    SELECT COUNT(*)=1 INTO memberExists  FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE FOR UPDATE;
    START TRANSACTION;
    IF memberExists THEN
      UPDATE projectMembers SET role=_role WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE;
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _memberId, 'member', 'setProjectRole', NULL, CAST(_role as char character set utf8));
    END IF;
    COMMIT;
    SELECT memberExists;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS createNode;
DELIMITER $$
CREATE PROCEDURE createNode(_accountId BINARY(16), _projectId BINARY(16), _parentId BINARY(16), _previousSiblingId BINARY(16), _nodeId BINARY(16), _isAbstract BOOLEAN, _name VARCHAR(250), _description VARCHAR(1250), _createdOn DATETIME, _totalRemainingTime BIGINT UNSIGNED, _isParallel BOOLEAN, _memberId BINARY(16))
BEGIN
  DECLARE _minimumRemainingTime BIGINT UNSIGNED DEFAULT _totalRemainingTime;
	DECLARE originalParentId BINARY(16) DEFAULT _parentId;
	DECLARE projectExists BOOLEAN DEFAULT FALSE;
  DECLARE parentExists BOOLEAN DEFAULT FALSE;
  DECLARE changeMade BOOLEAN DEFAULT FALSE;
	DECLARE previousSiblingExists BOOLEAN DEFAULT FALSE;
	DECLARE idVariable BINARY(16) DEFAULT NULL; #initialy used for holding nextSiblingIdToUse then used as a temporary parentId variable when looping up the tree
  DECLARE memberExistsAndIsActive BOOLEAN DEFAULT FALSE;
  #necessary values for traveling up the tree
  DECLARE currentIsParallel BOOLEAN DEFAULT FALSE;
  DECLARE currentMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;

  #assume parameters are already validated against themselves in the business logic

  START TRANSACTION;
    #validate parameters against the database, i.e. make sure parent/sibiling/member exist and are active and have sufficient privelages 
	#lock the project and lock the project member (member needs "for update" lock if remaining time is greater than zero, otherwise only "lock in share mode" is needed)
	SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
  IF _parentId=_projectId THEN
    SET parentExists=projectExists;
  ELSE
    SELECT COUNT(*)=1 INTO parentExists FROM nodes WHERE account=_accountId AND project=_projectId AND id=_parentId AND isAbstract=TRUE;
  END IF;
  IF _previousSiblingId IS NULL THEN
    SELECT firstChild INTO idVariable FROM nodes WHERE account=_accountId AND project=_projectId AND id=_parentId AND isAbstract=TRUE;
    SET previousSiblingExists=TRUE;
	ELSE
        SELECT COUNT(*)=1, nextSibling INTO previousSiblingExists, idVariable FROM nodes WHERE account=_accountId AND project=_projectId AND parent=_parentId AND id=_previousSiblingId;
  END IF;
  IF _memberId IS NOT NULL AND _totalRemainingTime <> 0 THEN
    SELECT COUNT(*)=1 INTO memberExistsAndIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE AND role < 2 FOR UPDATE; #less than 2 means 0->projectAdmin or 1->projectWriter
  ELSEIF _memberId IS NOT NULL AND _totalRemainingTime=0 THEN
    SELECT COUNT(*)=1 INTO memberExistsAndIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE AND role < 2 LOCK IN SHARE MODE;
  ELSE
    SET memberExistsAndIsActive=TRUE;
  END IF;
  IF _memberId IS NOT NULL AND _isAbstract THEN
    SET memberExistsAndIsActive=FALSE; #set this to false to fail the validation check so we dont create an invalid abstract node with an assigned member
  END IF;
  IF projectExists AND parentExists AND previousSiblingExists AND memberExistsAndIsActive THEN
    SET changeMade = TRUE;
		#write the node row
    if NOT _isAbstract THEN
      SET _isParallel=FALSE;
    END IF;
    INSERT INTO nodes (account,	project, id, parent, firstChild, nextSibling, isAbstract, name,	description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member) VALUES (_accountId,	_projectId, _nodeId, _parentId, NULL, idVariable, _isAbstract, _name, _description, _createdOn, _totalRemainingTime, 0, _minimumRemainingTime, 0, 0, 0, 0, _isParallel, _memberId);
    #update siblings and parent firstChild value if required
    IF _previousSiblingId IS NULL THEN #update parents firstChild
      UPDATE nodes SET firstChild=_nodeId WHERE account=_accountId AND project=_projectId AND id=_parentId;
    ELSE #update previousSiblings nextSibling value
      UPDATE nodes SET nextSibling=_nodeId WHERE account=_accountId AND project=_projectId AND id=_previousSiblingId;
    END IF;
    #update member if needed
    IF _memberId IS NOT NULL AND _totalRemainingTime <> 0 THEN
      UPDATE projectMembers SET totalRemainingTime=totalRemainingTime + _totalRemainingTime WHERE account=_accountId AND project=_projectId AND id=_memberId;
    END IF;
        
    WHILE _parentId IS NOT NULL DO
        
			#get values needed to update current node
			SELECT isParallel, minimumRemainingTime, parent INTO currentIsParallel, currentMinimumRemainingTime, idVariable FROM nodes WHERE account=_accountId AND project=_projectId AND id=_parentId;
            
        IF _totalRemainingTime=0 THEN #dont need to update time values
				IF _parentId <> originalParentId THEN #dont need to update child count
					UPDATE nodes SET descendantCount=descendantCount + 1 WHERE account=_accountId AND project=_projectId AND id=_parentId;
                ELSE #need to update child/descendant counts
					UPDATE nodes SET childCount=childCount + 1, descendantCount=descendantCount + 1 WHERE account=_accountId AND project=_projectId AND id=_parentId;
                END IF;
                #update _minimumRemainingTime to be correct value for next iteration
				SET _minimumRemainingTime=currentMinimumRemainingTime;
            ELSE #need to update time values
				#update _minimumRemainingTime to be correct value for current node
				IF currentIsParallel THEN
					IF _minimumRemainingTime < currentMinimumRemainingTime THEN #UPDATE total and minimum times and childCount and descendantCount
						SET _minimumRemainingTime=currentMinimumRemainingTime;
					END IF;
				ELSE
					SET _minimumRemainingTime=currentMinimumRemainingTime + _minimumRemainingTime;
				END IF;
                IF _parentId <> originalParentId THEN #dont need to update child count
					UPDATE nodes SET descendantCount=descendantCount + 1, totalRemainingTime=totalRemainingTime + _totalRemainingTime, minimumRemainingTime=_minimumRemainingTime WHERE account=_accountId AND project=_projectId AND id=_parentId;
                ELSE #need to update child/descendant counts
					UPDATE nodes SET childCount=childCount + 1, descendantCount=descendantCount + 1, totalRemainingTime=totalRemainingTime + _totalRemainingTime, minimumRemainingTime=_minimumRemainingTime WHERE account=_accountId AND project=_projectId AND id=_parentId;
                END IF;
            END IF;
            SET _parentId=idVariable;
        END WHILE;
		SELECT TRUE;
    END IF;
    SELECT changeMade;
    COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setNodeName;
DELIMITER $$
CREATE PROCEDURE setNodeName(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _name VARCHAR(250))
BEGIN
	UPDATE nodes SET name=_name WHERE account=_accountId AND project=_projectId AND id=_nodeId;
  IF _projectId=_nodeId THEN
    UPDATE projects SET name=_name WHERE account=_accountId AND id=_nodeId;
  ELSE
    UPDATE timeLogs SET nodeName=_name WHERE account=_accountId AND project=_projectId AND node=_nodeId;
  END IF;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setNodeIsParallel;
DELIMITER $$
CREATE PROCEDURE setNodeIsParallel(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _isParallel BOOLEAN)
BEGIN
	DECLARE projExists BOOLEAN DEFAULT FALSE;
  DECLARE nodeExists BOOLEAN DEFAULT FALSE;
  DECLARE changeMade BOOLEAN DEFAULT FALSE;
  DECLARE nextNode BINARY(16) DEFAULT NULL;
  DECLARE currentIsParallel BOOLEAN DEFAULT FALSE;
  DECLARE currentMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE preChangePreviousMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE postChangePreviousMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  START TRANSACTION;
  SELECT COUNT(*)=1 INTO projExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE; #set project lock to ensure data integrity
  IF projExists THEN
    SELECT COUNT(*)=1, parent, isParallel, minimumRemainingTime INTO nodeExists, nextNode, currentIsParallel, preChangePreviousMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId;
    IF nodeExists AND _isParallel <> currentIsParallel THEN #make sure we are making a change otherwise, no need to update anything
      SET changeMade = TRUE;
      IF preChangePreviousMinimumRemainingTime <> 0 THEN
        IF _isParallel THEN #setting isParallel to true
          SELECT MAX(minimumRemainingTime) INTO postChangePreviousMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND parent=_nodeId;
        ELSE #setting isParallel to false
          SELECT SUM(minimumRemainingTime) INTO postChangePreviousMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND parent=_nodeId;
        END IF;
        UPDATE nodes SET minimumRemainingTime=postChangePreviousMinimumRemainingTime, isParallel=_isParallel WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      END IF;

      SET _nodeId=nextNode;

      WHILE _nodeId IS NOT NULL AND preChangePreviousMinimumRemainingTime <> postChangePreviousMinimumRemainingTime DO
        #get values needed to update current node
        SELECT isParallel, minimumRemainingTime, parent INTO _isParallel, currentMinimumRemainingTime, nextNode FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId;
        IF _isParallel AND currentMinimumRemainingTime < postChangePreviousMinimumRemainingTime THEN
          UPDATE nodes SET minimumRemainingTime=postChangePreviousMinimumRemainingTime WHERE account=_accountId AND project=_projectId AND id=_nodeId;
        ELSEIF _isParallel AND currentMinimumRemainingTime = preChangePreviousMinimumRemainingTime THEN
          SELECT MAX(minimumRemainingTime) INTO postChangePreviousMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND parent=_nodeId;
          IF currentMinimumRemainingTime <> postChangePreviousMinimumRemainingTime THEN
            UPDATE nodes SET minimumRemainingTime=postChangePreviousMinimumRemainingTime WHERE account=_accountId AND project=_projectId AND id=_nodeId;
          END IF;
        ELSEIF NOT _isParallel THEN #setting isParallel to false
          UPDATE nodes SET minimumRemainingTime=minimumRemainingTime+postChangePreviousMinimumRemainingTime-preChangePreviousMinimumRemainingTime WHERE account=_accountId AND project=_projectId AND id=_nodeId;
          SET postChangePreviousMinimumRemainingTime=currentMinimumRemainingTime+postChangePreviousMinimumRemainingTime-preChangePreviousMinimumRemainingTime;
        ELSE #nochange to make
          SET postChangePreviousMinimumRemainingTime=currentMinimumRemainingTime;
        END IF;
        SET preChangePreviousMinimumRemainingTime=currentMinimumRemainingTime;
        SET _nodeId=nextNode;
      END WHILE;
    END IF;
  END IF;
  SELECT changeMade;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setNodeMember;
DELIMITER $$
CREATE PROCEDURE setNodeMember(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _memberId BINARY(16))
BEGIN
  DECLARE projExists BOOLEAN DEFAULT FALSE;
  DECLARE memberExistsAndIsActive BOOLEAN DEFAULT TRUE;
  DECLARE nodeExistsAndIsConcrete BOOLEAN DEFAULT FALSE;
  DECLARE changeMade BOOLEAN DEFAULT FALSE;
  DECLARE nodeTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE existingMemberTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE existingMember BINARY(16) DEFAULT NULL;
  DECLARE existingMemberExists BOOLEAN DEFAULT FALSE; #seems pointless but it's justa var to stick a value in when locking the row
  START TRANSACTION;
  SELECT COUNT(*)=1 INTO projExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE; #set project lock to ensure data integrity
  IF projExists THEN
    SELECT COUNT(*)=1, totalRemainingTime, member INTO nodeExistsAndIsConcrete, nodeTotalRemainingTime, existingMember FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId AND isAbstract=FALSE;
    IF (existingMember <> _memberId OR (existingMember IS NOT NULL AND _memberId IS NULL) OR (existingMember IS NULL AND _memberId IS NOT NULL)) AND nodeExistsAndIsConcrete THEN
      IF nodeTotalRemainingTime > 0 THEN
        IF _memberId IS NOT NULL THEN
          SELECT COUNT(*)=1 INTO memberExistsAndIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE AND role < 2 FOR UPDATE; #less than 2 means 0->projectAdmin or 1->projectWriter
          IF memberExistsAndIsActive THEN
            UPDATE projectMembers SET totalRemainingTime=totalRemainingTime+nodeTotalRemainingTime WHERE account=_accountId AND project=_projectId AND id=_memberId;
          END IF;
        END IF;
        IF existingMember IS NOT NULL THEN
          SELECT COUNT(*)=1, totalRemainingTime INTO existingMemberExists, existingMemberTotalRemainingTime FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=existingMember FOR UPDATE; #less than 2 means 0->projectAdmin or 1->projectWriter
          IF existingMemberTotalRemainingTime >= nodeTotalRemainingTime THEN
            UPDATE projectMembers SET totalRemainingTime=totalRemainingTime-nodeTotalRemainingTime WHERE account=_accountId AND project=_projectId AND id=existingMember;
          END IF;
        END IF;
      ELSE
        IF _memberId IS NOT NULL THEN
          SELECT COUNT(*)=1 INTO memberExistsAndIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE AND role < 2 LOCK IN SHARE MODE; #less than 2 means 0->projectAdmin or 1->projectWriter
        END IF;
      END IF;
      IF memberExistsAndIsActive THEN
        UPDATE nodes SET member=_memberId WHERE account=_accountId AND project=_projectId AND id=_nodeId;
        SET changeMade = TRUE;
      END IF;
    END IF;
  END IF;
  SELECT changeMade;
  COMMIT;
END;
$$
DELIMITER ;

## Pass NULL in _timeRemaining to not set a new TotalTimeRemaining value, pass NULL or zero to _duration to not log time
DROP PROCEDURE IF EXISTS setTimeRemainingAndOrLogTime;
DELIMITER $$
CREATE PROCEDURE setTimeRemainingAndOrLogTime(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _timeRemaining BIGINT UNSIGNED, _myId BINARY(16), _loggedOn DATETIME, _duration BIGINT UNSIGNED, _note VARCHAR(250))
BEGIN
  DECLARE projExists BOOLEAN DEFAULT FALSE;
  DECLARE nodeExists BOOLEAN DEFAULT FALSE;
  DECLARE memberExists BOOLEAN DEFAULT FALSE;
  DECLARE nodeName VARCHAR(250) DEFAULT '';
  DECLARE changeMade BOOLEAN DEFAULT FALSE;
  DECLARE existingMemberExists BOOLEAN DEFAULT FALSE;
  DECLARE existingMember BINARY(16) DEFAULT NULL;
  DECLARE nextNode BINARY(16) DEFAULT NULL;
  DECLARE currentIsParallel BOOLEAN DEFAULT FALSE;
  DECLARE currentMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE preChangeOriginalMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE preChangePreviousMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE postChangePreviousMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  START TRANSACTION;
  SELECT COUNT(*)=1 INTO projExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
  IF projExists THEN
    SELECT COUNT(*)=1, name, member, parent, totalRemainingTime INTO nodeExists, nodeName, existingMember, nextNode, preChangeOriginalMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId AND isAbstract=FALSE;
    SET preChangePreviousMinimumRemainingTime=preChangeOriginalMinimumRemainingTime;
    IF _timeRemaining IS NULL THEN
      SET _timeRemaining=preChangeOriginalMinimumRemainingTime;
    END IF;
    IF _duration IS NULL THEN
      SET _duration=0;
    END IF;
    IF nodeExists AND (preChangeOriginalMinimumRemainingTime <> _timeRemaining OR _duration > 0) THEN
      IF existingMember IS NOT NULL AND preChangeOriginalMinimumRemainingTime <> _timeRemaining THEN
        SELECT COUNT(*)=1 INTO existingMemberExists FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=existingMember FOR UPDATE;
        UPDATE projectMembers SET totalRemainingTime=totalRemainingTime+_timeRemaining-preChangeOriginalMinimumRemainingTime WHERE account=_accountId AND project=_projectId AND id=existingMember;
      END IF;

      IF _myId IS NOT NULL AND _duration > 0 THEN
        IF existingMember IS NULL OR existingMember <> _myId THEN
          SELECT COUNT(*)=1 INTO memberExists FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_myId FOR UPDATE;
        ELSE #we already set the lock and proved it exists as it is the member assigned to the task itself
          SET memberExists=TRUE;
        END IF;
        IF memberExists THEN
          INSERT INTO timeLogs (account, project, node, member, loggedOn, nodeName, duration, note) VALUES (_accountId, _projectId, _nodeId, _myId, _loggedOn, nodeName, _duration, _note);
          UPDATE projectMembers SET totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_myId;
        END IF;
      END IF;

      UPDATE nodes SET totalRemainingTime=_timeRemaining, minimumRemainingTime=_timeRemaining, totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      SET _nodeId=nextNode;
      SET postChangePreviousMinimumRemainingTime=_timeRemaining;
      WHILE _nodeId IS NOT NULL DO
        IF preChangePreviousMinimumRemainingTime <> postChangePreviousMinimumRemainingTime THEN #updating total and minimum remaining times
          #get values needed to update current node
          SELECT isParallel, minimumRemainingTime, parent INTO currentIsParallel, currentMinimumRemainingTime, nextNode FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId;
          IF currentIsParallel AND currentMinimumRemainingTime < postChangePreviousMinimumRemainingTime THEN
            UPDATE nodes SET totalRemainingTime=totalRemainingTime+_timeRemaining-preChangeOriginalMinimumRemainingTime, minimumRemainingTime=postChangePreviousMinimumRemainingTime, totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_nodeId;
          ELSEIF currentIsParallel AND currentMinimumRemainingTime = preChangePreviousMinimumRemainingTime THEN
            SELECT MAX(minimumRemainingTime) INTO postChangePreviousMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND parent=_nodeId;
            IF currentMinimumRemainingTime <> postChangePreviousMinimumRemainingTime THEN
              UPDATE nodes SET totalRemainingTime=totalRemainingTime+_timeRemaining-preChangeOriginalMinimumRemainingTime, minimumRemainingTime=postChangePreviousMinimumRemainingTime, totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_nodeId;
            END IF;
          ELSEIF (NOT currentIsParallel) THEN
            UPDATE nodes SET totalRemainingTime=totalRemainingTime+_timeRemaining-preChangeOriginalMinimumRemainingTime, minimumRemainingTime=minimumRemainingTime+postChangePreviousMinimumRemainingTime-preChangePreviousMinimumRemainingTime, totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_nodeId;
            SET postChangePreviousMinimumRemainingTime=currentMinimumRemainingTime+postChangePreviousMinimumRemainingTime-preChangePreviousMinimumRemainingTime;
          ELSE #nochange to minimum time to make
            UPDATE nodes SET totalRemainingTime=totalRemainingTime+_timeRemaining-preChangeOriginalMinimumRemainingTime, totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_nodeId;
            SET postChangePreviousMinimumRemainingTime=currentMinimumRemainingTime;
          END IF;
          SET preChangePreviousMinimumRemainingTime=currentMinimumRemainingTime;
        ELSE #only updating total remaining time
          SELECT parent INTO nextNode FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId;
          UPDATE nodes SET totalRemainingTime=totalRemainingTime+_timeRemaining-preChangeOriginalMinimumRemainingTime, totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_nodeId;
        END IF;
        SET _nodeId=nextNode;
      END WHILE;
      SET changeMade = TRUE;
    END IF;
  END IF;
  SELECT changeMade;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS moveNode;
DELIMITER $$
CREATE PROCEDURE moveNode(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _newParentId BINARY(16), _newPreviousSibling BINARY(16))
BEGIN
  DECLARE projExists BOOLEAN DEFAULT FALSE;
  DECLARE nodeExists BOOLEAN DEFAULT FALSE;
  DECLARE newParentExists BOOLEAN DEFAULT TRUE;
  DECLARE newParentFirstChildId BINARY(16) DEFAULT NULL;
  DECLARE newPreviousSiblingExists BOOLEAN DEFAULT TRUE;
  DECLARE newNextSiblingId BINARY(16) DEFAULT NULL;
  DECLARE originalParentId BINARY(16) DEFAULT NULL;
  DECLARE originalParentFirstChildId BINARY(16) DEFAULT NULL;
  DECLARE originalNextSiblingId BINARY(16) DEFAULT NULL;
  DECLARE originalPreviousSiblingId BINARY(16) DEFAULT NULL;
  DECLARE changeMade BOOLEAN DEFAULT FALSE;
  START TRANSACTION;
  IF _newParentId IS NOT NULL AND _projectId <> _nodeId THEN #check newParent is not null AND _nodeId <> _projectId or this call is invalid
    SELECT COUNT(*)=1 INTO projExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
    IF projExists THEN
      SELECT COUNT(*)=1, parent, nextSibling INTO nodeExists, originalParentId, originalNextSiblingId FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      IF nodeExists THEN
        IF _newParentId = originalParentId THEN #most moves will be just moving within a single parent node, this is a very simple and efficient move to make as we dont need to walk up the tree updating child/descendant counts or total/minimum remaining time values
          SELECT firstChild INTO originalParentFirstChildId FROM nodes WHERE account=_accountId AND project=_projectId AND id=originalParentId;
          IF _newPreviousSibling IS NULL AND originalParentFirstChildId <> _nodeId  THEN #moving to firstChild position, and it is an actual change
            SELECT id INTO originalPreviousSiblingId FROM nodes WHERE account=_accountId AND project=_projectId AND nextSibling=_nodeId;
            UPDATE nodes SET nextSibling=originalParentFirstChildId WHERE account=_accountId AND project=_projectId AND id=_nodeId;
            UPDATE nodes SET nextSibling=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalPreviousSiblingId;
            UPDATE nodes SET firstChild=_nodeId WHERE account=_accountId AND project=_projectId AND id=originalParentId;
            SET changeMade=TRUE;
          ELSEIF _newPreviousSibling IS NOT NULL THEN #moving to some other position NOT firstChild
            SELECT COUNT(*)=1, nextSibling INTO newPreviousSiblingExists, newNextSiblingId FROM nodes WHERE account=_accountId AND project=_projectId AND id=_newPreviousSibling AND parent=originalParentId;
            IF newPreviousSiblingExists AND ((newNextSiblingId IS NULL AND originalNextSiblingId IS NOT NULL) OR (newNextSiblingId IS NOT NULL AND originalNextSiblingId IS NULL) OR (newNextSiblingId IS NOT NULL AND originalNextSiblingId IS NOT NULL AND newNextSiblingId <> originalNextSiblingId)) THEN
              UPDATE nodes SET nextSibling=newNextSiblingId WHERE account=_accountId AND project=_projectId AND id=_nodeId;
              UPDATE nodes SET nextSibling=_nodeId WHERE account=_accountId AND project=_projectId AND id=_newPreviousSibling;
              IF originalParentFirstChildId = _nodeId AND originalNextSiblingId IS NOT NULL THEN #moving the first child node
                UPDATE nodes SET firstChild=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalParentId;
              ELSE #moving a none first child node
                SELECT id INTO originalPreviousSiblingId FROM nodes WHERE account=_accountId AND project=_projectId AND nextSibling=_nodeId;
                IF originalPreviousSiblingId IS NOT NULL THEN
                  UPDATE nodes SET nextSibling=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalPreviousSiblingId;
                END IF;
              END IF;
              SET changeMade=TRUE;
            END IF;
          END IF;
        ELSE #this is the expensive complex move to make, to simplify we do not work out a shared ancestor node and perform operations up to that node, we full remove the aggregated values all the way to the project node, then add them back in in the new location, this may result in more processing, but should still be efficient and simplify the code logic below.
			    SET projExists=FALSE;
        END IF;
      END IF;
    END IF;
  END IF;
  SELECT changeMade;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteNode;
DELIMITER $$
CREATE PROCEDURE deleteNode(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16))
BEGIN
	#TODO
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS getNodes;
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