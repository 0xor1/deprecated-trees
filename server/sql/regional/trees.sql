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
  isArchived BOOL NOT NULL,
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
CREATE PROCEDURE setPublicProjectsEnabled(_accountId BINARY(16), _myId BINARY(16), _enabled BOOL)
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
    DECLARE memberExists BOOL DEFAULT FALSE;
    SELECT COUNT(*)=1 INTO memberExists  FROM accountMembers WHERE account=_accountId AND id=_memberId AND isActive=TRUE FOR UPDATE;
    START TRANSACTION;
    IF memberExists THEN
      UPDATE accountMembers SET role=_role WHERE account=_accountId AND id=_memberId AND isActive=TRUE;
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _memberId, 'member', 'setRole', NULL, CAST(_role as char character set utf8));
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
	DECLARE projMemberExists BOOL DEFAULT FALSE;
  DECLARE projMemberIsActive BOOL DEFAULT FALSE;
	DECLARE accMemberName VARCHAR(50) DEFAULT '';
	DECLARE accMemberDisplayName VARCHAR(100) DEFAULT NULL;
  DECLARE changeMade BOOL DEFAULT FALSE;
  SELECT COUNT(*)=1, isActive INTO projMemberExists, projMemberIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_id;
  IF projMemberExists AND projMemberIsActive=false THEN #setting previous member back to active, still need to check if they are an active account member
    IF (SELECT COUNT(*) FROM accountMembers WHERE account=_accountId AND id=_id AND isActive=true) THEN #if active account member then add them to the project
      UPDATE projectMembers SET role=_role, isActive=true WHERE account=_accountId AND project=_projectId AND id=_id;
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _id, 'member', 'added', NULL, NULL);
      SET changeMade = true;
    END IF;
	ELSEIF NOT projMemberExists THEN #adding new project member, need to check if they are active account member
		START TRANSACTION;
			SELECT name, displayName INTO accMemberName, accMemberDisplayName FROM accountMembers WHERE account=_accountId AND id=_id AND isActive=true LOCK IN SHARE MODE;
			IF accMemberName IS NOT NULL AND accMemberName <> '' THEN #if active account member then add them to the project
				INSERT INTO projectMembers (account, project, id, name, displayName, isActive, totalRemainingTime, totalLoggedTime, role) VALUES (_accountId, _projectId, _id, accMemberName, accMemberDisplayName, true, 0, 0, _role);
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _id, 'member', 'added', NULL, NULL);
        SET changeMade = true;
			END IF;
    COMMIT;
  END IF;
  SELECT changeMade;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setProjectMemberInactive;
DELIMITER $$
CREATE PROCEDURE setProjectMemberInactive(_accountId BINARY(16), _projectId BINARY(16), _myId BINARY(16), _id BINARY(16))
BEGIN
  DECLARE projExists BOOL DEFAULT FALSE;
  DECLARE projMemberExists BOOL DEFAULT FALSE;
  DECLARE changeMade BOOL DEFAULT FALSE;

  START TRANSACTION;
    SELECT COUNT(*)=1 INTO projExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
    IF projExists THEN
      SELECT COUNT(*)=1 INTO projMemberExists FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_id AND isActive=true FOR UPDATE;
      IF projMemberExists THEN
        UPDATE nodes SET member=NULL WHERE account=_accountId AND project=_projectId AND member=_id;
        UPDATE projectMembers SET totalRemainingTime=0 WHERE account=_accountId AND project=_projectId AND id=_id;
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _id, 'member', 'removed', NULL, NULL);
        SET changeMade = true;
      END IF;
    END IF;
  COMMIT;
  SELECT changeMade;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setProjectMemberRole;
DELIMITER $$
CREATE PROCEDURE setProjectMemberRole(_accountId BINARY(16), _projectId BINARY(16), _myId BINARY(16), _memberId BINARY(16), _role TINYINT UNSIGNED)
  BEGIN
    DECLARE memberExists BOOL DEFAULT FALSE;
    SELECT COUNT(*)=1 INTO memberExists  FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE FOR UPDATE;
    START TRANSACTION;
    IF memberExists THEN
      UPDATE projectMembers SET role=_role WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE;
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _memberId, 'member', 'setRole', NULL, CAST(_role as char character set utf8));
    END IF;
    COMMIT;
    SELECT memberExists;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS createNode;
DELIMITER $$
CREATE PROCEDURE createNode(_accountId BINARY(16), _projectId BINARY(16), _parentId BINARY(16), _myId BINARY(16), _previousSiblingId BINARY(16), _nodeId BINARY(16), _isAbstract BOOL, _name VARCHAR(250), _description VARCHAR(1250), _createdOn DATETIME, _totalRemainingTime BIGINT UNSIGNED, _isParallel BOOL, _memberId BINARY(16))
BEGIN
	DECLARE projectExists BOOL DEFAULT FALSE;
  DECLARE parentExists BOOL DEFAULT FALSE;
	DECLARE previousSiblingExists BOOL DEFAULT FALSE;
	DECLARE nextSiblingToUse BINARY(16) DEFAULT NULL;
  DECLARE memberExistsAndIsActive BOOL DEFAULT FALSE;
  DECLARE changeMade BOOL DEFAULT FALSE;

  #assume parameters are already validated against themselves in the business logic

  START TRANSACTION;
  #validate parameters against the database, i.e. make sure parent/sibiling/member exist and are active and have sufficient privileges
	#lock the project and lock the project member (member needs "for update" lock if remaining time is greater than zero, otherwise only "lock in share mode" is needed)
	SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
  IF _parentId=_projectId THEN
    SET parentExists=projectExists;
  ELSE
    SELECT COUNT(*)=1 INTO parentExists FROM nodes WHERE account=_accountId AND project=_projectId AND id=_parentId AND isAbstract=TRUE;
  END IF;
  IF _previousSiblingId IS NULL THEN
    SELECT firstChild INTO nextSiblingToUse FROM nodes WHERE account=_accountId AND project=_projectId AND id=_parentId AND isAbstract=TRUE;
    SET previousSiblingExists=TRUE;
	ELSE
    SELECT COUNT(*)=1, nextSibling INTO previousSiblingExists, nextSiblingToUse FROM nodes WHERE account=_accountId AND project=_projectId AND parent=_parentId AND id=_previousSiblingId;
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
    SET changeMade=TRUE;
		#write the node row
    if NOT _isAbstract THEN
      SET _isParallel=FALSE;
    END IF;
    INSERT INTO nodes (account,	project, id, parent, firstChild, nextSibling, isAbstract, name,	description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member) VALUES (_accountId,	_projectId, _nodeId, _parentId, NULL, nextSiblingToUse, _isAbstract, _name, _description, _createdOn, _totalRemainingTime, 0, _totalRemainingTime, 0, 0, 0, 0, _isParallel, _memberId);
    INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'node', 'created', _name, NULL);
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

    CALL _setAncestralChainAggregateValuesFromNode(_accountId, _projectId, _parentId);
  END IF;
  SELECT changeMade;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setNodeName;
DELIMITER $$
CREATE PROCEDURE setNodeName(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _myId BINARY(16), _name VARCHAR(250))
  BEGIN
    UPDATE nodes SET name=_name WHERE account=_accountId AND project=_projectId AND id=_nodeId;
    IF _projectId=_nodeId THEN
      UPDATE projects SET name=_name WHERE account=_accountId AND id=_nodeId;
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'project', 'setName', NULL, _name);
      UPDATE accountActivities SET itemName=_name WHERE account=_accountId AND item=_nodeId;
    ELSE
      UPDATE timeLogs SET nodeName=_name WHERE account=_accountId AND project=_projectId AND node=_nodeId;
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'node', 'setName', NULL, _name);
      UPDATE projectActivities SET itemName=_name WHERE account=_accountId AND project=_projectId AND item=_nodeId;
    END IF;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setNodeDescription;
DELIMITER $$
CREATE PROCEDURE setNodeDescription(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _myId BINARY(16), _description VARCHAR(250))
  BEGIN
    UPDATE nodes SET description=_description WHERE account=_accountId AND project=_projectId AND id=_nodeId;
    INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'node', 'setDescription', NULL, _description);
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setNodeIsParallel;
DELIMITER $$
CREATE PROCEDURE setNodeIsParallel(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _myId BINARY(16), _isParallel BOOL)
BEGIN
	DECLARE projExists BOOL DEFAULT FALSE;
  DECLARE nodeExists BOOL DEFAULT FALSE;
  DECLARE nextNode BINARY(16) DEFAULT NULL;
  DECLARE currentIsParallel BOOL DEFAULT FALSE;
  DECLARE currentMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE changeMade BOOL DEFAULT FALSE;
  START TRANSACTION;
  SELECT COUNT(*)=1 INTO projExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE; #set project lock to ensure data integrity
  IF projExists THEN
    SELECT COUNT(*)=1, parent, isParallel, minimumRemainingTime INTO nodeExists, nextNode, currentIsParallel, currentMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId;
    IF nodeExists AND _isParallel <> currentIsParallel THEN #make sure we are making a change otherwise, no need to update anything
      SET changeMade = TRUE;
      IF _isParallel THEN
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'node', 'setIsParallel', NULL, 'true');
      ELSE
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'node', 'setIsParallel', NULL, 'false');
      END IF;
      UPDATE nodes SET isParallel=_isParallel WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      IF currentMinimumRemainingTime <> 0 THEN
        CALL _setAncestralChainAggregateValuesFromNode(_accountId, _projectId, _nodeId);
      END IF;
    END IF;
  END IF;
  SELECT changeMade;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setNodeMember;
DELIMITER $$
CREATE PROCEDURE setNodeMember(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _myId BINARY(16), _memberId BINARY(16))
BEGIN
  DECLARE projExists BOOL DEFAULT FALSE;
  DECLARE memberExistsAndIsActive BOOL DEFAULT TRUE;
  DECLARE nodeExistsAndIsConcrete BOOL DEFAULT FALSE;
  DECLARE changeMade BOOL DEFAULT FALSE;
  DECLARE nodeTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE existingMemberTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE existingMember BINARY(16) DEFAULT NULL;
  DECLARE existingMemberExists BOOL DEFAULT FALSE; #seems pointless but it's justa var to stick a value in when locking the row
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
        IF _memberId IS NOT NULL THEN
          INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'node', 'setMember', NULL, LOWER(HEX(_memberId)));
        ELSE
          INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'node', 'setMember', NULL, NULL);
        END IF;
      END IF;
    END IF;
  END IF;
  SELECT changeMade;
  COMMIT;
END;
$$
DELIMITER ;

## Pass NULL in _timeRemaining to not set a new TotalTimeRemaining value, pass NULL or zero to _duration to not log time
DROP PROCEDURE IF EXISTS setRemainingTimeAndOrLogTime;
DELIMITER $$
CREATE PROCEDURE setRemainingTimeAndOrLogTime(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _myId bINARY(16), _timeRemaining BIGINT UNSIGNED, _loggedOn DATETIME, _duration BIGINT UNSIGNED, _note VARCHAR(250))
BEGIN
  DECLARE projExists BOOL DEFAULT FALSE;
  DECLARE nodeExists BOOL DEFAULT FALSE;
  DECLARE memberExists BOOL DEFAULT FALSE;
  DECLARE nodeName VARCHAR(250) DEFAULT '';
  DECLARE existingMemberExists BOOL DEFAULT FALSE;
  DECLARE existingMember BINARY(16) DEFAULT NULL;
  DECLARE nextNode BINARY(16) DEFAULT NULL;
  DECLARE originalMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE changeMade BOOL DEFAULT FALSE;
  START TRANSACTION;
  SELECT COUNT(*)=1 INTO projExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
  IF projExists THEN
    SELECT COUNT(*)=1, name, member, parent, totalRemainingTime INTO nodeExists, nodeName, existingMember, nextNode, originalMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId AND isAbstract=FALSE;
    IF _timeRemaining IS NULL THEN
      SET _timeRemaining= originalMinimumRemainingTime;
    END IF;
    IF _duration IS NULL THEN
      SET _duration=0;
    END IF;
    IF nodeExists AND (originalMinimumRemainingTime <> _timeRemaining OR _duration > 0) THEN
      SET changeMade = TRUE;
      IF existingMember IS NOT NULL AND originalMinimumRemainingTime <> _timeRemaining THEN
        SELECT COUNT(*)=1 INTO existingMemberExists FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=existingMember FOR UPDATE;
        UPDATE projectMembers SET totalRemainingTime=totalRemainingTime+_timeRemaining - originalMinimumRemainingTime WHERE account = _accountId AND project = _projectId AND id = existingMember;
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
          INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'node', 'loggedTime', nodeName, _note);
        END IF;
      END IF;

      UPDATE nodes SET totalRemainingTime=_timeRemaining, minimumRemainingTime=_timeRemaining, totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      IF _timeRemaining <> originalMinimumRemainingTime THEN
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, ADDTIME(UTC_TIMESTAMP(6), '0:0:0.000001'), _myId, _nodeId, 'node', 'setRemainingTime', nodeName, CAST(_timeRemaining as char character set utf8));
      END IF;
      CALL _setAncestralChainAggregateValuesFromNode(_accountId, _projectId, nextNode);
    END IF;
  END IF;
  SELECT changeMade;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS moveNode;
DELIMITER $$
CREATE PROCEDURE moveNode(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16), _newParentId BINARY(16), _myId BINARY(16), _newPreviousSiblingId BINARY(16))
BEGIN
  DECLARE projExists BOOL DEFAULT FALSE;
  DECLARE nodeExists BOOL DEFAULT FALSE;
  DECLARE nodeName VARCHAR(250) DEFAULT '';
  DECLARE newParentExists BOOL DEFAULT TRUE;
  DECLARE newParentFirstChildId BINARY(16) DEFAULT NULL;
  DECLARE newPreviousSiblingExists BOOL DEFAULT TRUE;
  DECLARE newNextSiblingId BINARY(16) DEFAULT NULL;
  DECLARE originalParentId BINARY(16) DEFAULT NULL;
  DECLARE originalParentFirstChildId BINARY(16) DEFAULT NULL;
  DECLARE originalNextSiblingId BINARY(16) DEFAULT NULL;
  DECLARE originalPreviousSiblingId BINARY(16) DEFAULT NULL;
  DECLARE originalTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE originalTotalLoggedTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE originalMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE changeMade BOOL DEFAULT FALSE;
  START TRANSACTION;
  IF _newParentId IS NOT NULL AND _projectId <> _nodeId AND _newParentId <> _nodeId THEN #check newParent is not null AND _nodeId is not the project or the new parent
    SELECT COUNT(*)=1 INTO projExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
    IF projExists THEN
      SELECT COUNT(*)=1, parent, nextSibling, name, totalRemainingTime, totalLoggedTime, minimumRemainingTime INTO nodeExists, originalParentId, originalNextSiblingId, nodeName, originalTotalRemainingTime, originalTotalLoggedTime, originalMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      IF nodeExists THEN
        SELECT firstChild INTO originalParentFirstChildId FROM nodes WHERE account=_accountId AND project=_projectId AND id=originalParentId;
        IF _newPreviousSiblingId IS NOT NULL THEN
          SELECT COUNT(*)=1, nextSibling INTO newPreviousSiblingExists, newNextSiblingId FROM nodes WHERE account=_accountId AND project=_projectId AND id=_newPreviousSiblingId AND parent=originalParentId;
        ELSE
          SET newPreviousSiblingExists=TRUE;
        END IF;
        IF newPreviousSiblingExists THEN
          IF _newParentId = originalParentId THEN #most moves will be just moving within a single parent node, this is a very simple and efficient move to make as we dont need to walk up the tree updating child/descendant counts or total/minimum/logged times
            IF _newPreviousSiblingId IS NULL AND originalParentFirstChildId <> _nodeId  THEN #moving to firstChild position, and it is an actual change
              SELECT id INTO originalPreviousSiblingId FROM nodes WHERE account=_accountId AND project=_projectId AND nextSibling=_nodeId;
              UPDATE nodes SET nextSibling=originalParentFirstChildId WHERE account=_accountId AND project=_projectId AND id=_nodeId;
              UPDATE nodes SET nextSibling=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalPreviousSiblingId;
              UPDATE nodes SET firstChild=_nodeId WHERE account=_accountId AND project=_projectId AND id=originalParentId;
              SET changeMade = TRUE;
            ELSEIF _newPreviousSiblingId IS NOT NULL AND ((newNextSiblingId IS NULL AND originalNextSiblingId IS NOT NULL) OR (newNextSiblingId IS NOT NULL AND newNextSiblingId <> _nodeId AND originalNextSiblingId IS NULL) OR (newNextSiblingId IS NOT NULL AND newNextSiblingId <> _nodeId AND originalNextSiblingId IS NOT NULL AND newNextSiblingId <> originalNextSiblingId)) THEN
              IF originalParentFirstChildId = _nodeId AND originalNextSiblingId IS NOT NULL THEN #moving the first child node
                UPDATE nodes SET firstChild=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalParentId;
              ELSE #moving a none first child node
                SELECT id INTO originalPreviousSiblingId FROM nodes WHERE account=_accountId AND project=_projectId AND nextSibling=_nodeId;
                IF originalPreviousSiblingId IS NOT NULL THEN
                  UPDATE nodes SET nextSibling=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalPreviousSiblingId;
                END IF;
              END IF;
              UPDATE nodes SET nextSibling=newNextSiblingId WHERE account=_accountId AND project=_projectId AND id=_nodeId;
              UPDATE nodes SET nextSibling=_nodeId WHERE account=_accountId AND project=_projectId AND id=_newPreviousSiblingId;
              SET changeMade = TRUE;
            END IF;
          ELSE # this is the expensive complex move to make, to simplify we do not work out a shared ancestor node and perform operations up to that node
               # we fully remove the aggregated values all the way to the project node, then add them back in in the new location, this may result in more processing,
               # but should still be efficient and simplify the code logic below.
            SELECT COUNT(*)=1, firstChild INTO newParentExists, newParentFirstChildId FROM nodes WHERE account=_accountId AND project=_projectId AND id=_newParentId;
            IF newParentExists THEN
              #move the node
              #remove from original location
              IF originalParentFirstChildId = _nodeId THEN #removing from first child position
                UPDATE nodes SET firstChild=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalParentId;
              ELSE #removing from none first child position
                SELECT id INTO originalPreviousSiblingId FROM nodes WHERE account=_accountId AND project=_projectId AND nextSibling=_nodeId;
                UPDATE nodes SET nextSibling=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalPreviousSiblingId;
              END IF;
              ##remove parent id so it isn't included in update in setAncestralChainAggregateValuesFromNode call
              UPDATE nodes SET parent=NULL WHERE account=_accountId AND project=_projectId AND id=_nodeId;
              CALL _setAncestralChainAggregateValuesFromNode(_accountId, _projectId, originalParentId);

              #now add in the new position
              IF _newPreviousSiblingId IS NULL THEN #moving to firstChild position
                UPDATE nodes SET firstChild=_nodeId WHERE account=_accountId AND project=_projectId AND id=_newParentId;
                UPDATE nodes SET parent=_newParentId, nextSibling=newParentFirstChildId WHERE account=_accountId AND project=_projectId AND id=_nodeId;
              ELSE #moving to none firstChild position
                UPDATE nodes SET nextSibling=_nodeId WHERE account=_accountId AND project=_projectId AND id=_newPreviousSiblingId;
                UPDATE nodes SET parent=_newParentId, nextSibling=newNextSiblingId WHERE account=_accountId AND project=_projectId AND id=_nodeId;
              END IF;
              CALL _setAncestralChainAggregateValuesFromNode(_accountId, _projectId, _newParentId);
              SET changeMade=TRUE;
            END IF;
          END IF;
        END IF;
      END IF;
    END IF;
  END IF;
  IF changeMade THEN
    INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, newValue) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _nodeId, 'node', 'moved', nodeName, NULL);
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

#!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!#
#********************************MAGIC PROCEDURE WARNING*********************************#
# THIS PROCEDURE MUST ONLY BE CALLED INTERNALLY BY THE ABOVE STORED PROCEDURES THAT HAVE #
# SET THEIR OWN TRANSACTIONS AND PROJECTID LOCKS AND HAVE VALIDATED ALL INPUT PARAMS.    #                                                                                                  #
#!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!#
DROP PROCEDURE IF EXISTS _setAncestralChainAggregateValuesFromNode;
DELIMITER $$
CREATE PROCEDURE _setAncestralChainAggregateValuesFromNode(_accountId BINARY(16), _projectId BINARY(16), _nodeId BINARY(16))
  BEGIN
    DECLARE originalTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE originalTotalLoggedTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE currentMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE originalChildCount BIGINT UNSIGNED DEFAULT 0;
    DECLARE originalDescendantCount BIGINT UNSIGNED DEFAULT 0;
    DECLARE currentIsParallel BOOL DEFAULT FALSE;
    DECLARE nextNode BINARY(16) DEFAULT NULL;
    DECLARE totalRemainingTimeChange BIGINT UNSIGNED DEFAULT 0;
    DECLARE totalLoggedTimeChange BIGINT UNSIGNED DEFAULT 0;
    DECLARE preChangeMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE postChangeMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE newChildCount BIGINT UNSIGNED DEFAULT 0;
    DECLARE descendantCountChange BIGINT UNSIGNED DEFAULT 0;
    DECLARE totalRemainingTimeChangeIsPositive BOOL DEFAULT TRUE;
    DECLARE totalLoggedTimeChangeIsPositive BOOL DEFAULT TRUE;
    DECLARE descendantCountChangeIsPositive BOOL DEFAULT TRUE;

    SELECT totalRemainingTime, totalLoggedTime, minimumRemainingTime, childCount, descendantCount, isParallel, parent INTO originalTotalRemainingTime, originalTotalLoggedTime, preChangeMinimumRemainingTime, originalChildCount, originalDescendantCount, currentIsParallel, nextNode FROM nodes WHERE account = _accountId AND project = _projectId AND id = _nodeId;
    IF currentIsParallel THEN
      SELECT SUM(totalRemainingTime), SUM(totalLoggedTime), MAX(minimumRemainingTime), COUNT(*), SUM(descendantCount)INTO totalRemainingTimeChange, totalLoggedTimeChange, postChangeMinimumRemainingTime, newChildCount, descendantCountChange FROM nodes WHERE account = _accountId AND project = _projectId AND parent = _nodeId;
    ELSE                                                   #this is the only difference#
      SELECT SUM(totalRemainingTime), SUM(totalLoggedTime), SUM(minimumRemainingTime), COUNT(*), SUM(descendantCount)INTO totalRemainingTimeChange, totalLoggedTimeChange, postChangeMinimumRemainingTime, newChildCount, descendantCountChange FROM nodes WHERE account = _accountId AND project = _projectId AND parent = _nodeId;
    END IF;
    SET descendantCountChange = descendantCountChange + newChildCount;

    #the first node updated is special, it could have had a new child added or removed from it, so the childCount can be updated, no other ancestor will have the childCount updated
    UPDATE nodes SET totalRemainingTime = totalRemainingTimeChange, totalLoggedTime = totalLoggedTimeChange, minimumRemainingTime = postChangeMinimumRemainingTime, childCount = newChildCount, descendantCount = descendantCountChange WHERE account = _accountId AND project = _projectId AND id = _nodeId;

    IF totalRemainingTimeChange >= originalTotalRemainingTime THEN
      SET totalRemainingTimeChange = totalRemainingTimeChange - originalTotalRemainingTime;
    ELSE
      SET totalRemainingTimeChange = originalTotalRemainingTime - totalRemainingTimeChange;
      SET totalRemainingTimeChangeIsPositive = FALSE;
    END IF;

    IF totalLoggedTimeChange >= originalTotalLoggedTime THEN
      SET totalLoggedTimeChange = totalLoggedTimeChange - originalTotalLoggedTime;
    ELSE
      SET totalLoggedTimeChange = originalTotalLoggedTime - totalLoggedTimeChange;
      SET totalLoggedTimeChangeIsPositive = FALSE;
    END IF;

    IF descendantCountChange >= originalDescendantCount THEN
      SET descendantCountChange = descendantCountChange - originalDescendantCount;
    ELSE
      SET descendantCountChange = originalDescendantCount - descendantCountChange;
      SET descendantCountChangeIsPositive = FALSE;
    END IF;

    SET _nodeId= nextNode;

    WHILE _nodeId IS NOT NULL AND (totalRemainingTimeChange > 0 OR totalLoggedTimeChange > 0 OR preChangeMinimumRemainingTime <> postChangeMinimumRemainingTime OR descendantCountChange > 0) DO
      IF preChangeMinimumRemainingTime <> postChangeMinimumRemainingTime THEN #updating minimumRemainingTime and others
        #get values needed to update current node
        SELECT isParallel, minimumRemainingTime, parent INTO currentIsParallel, currentMinimumRemainingTime, nextNode FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId;
        IF currentIsParallel AND currentMinimumRemainingTime < postChangeMinimumRemainingTime THEN
          SET postChangeMinimumRemainingTime = postChangeMinimumRemainingTime; #pointless assignment but this if case is necessary
        ELSEIF currentIsParallel AND currentMinimumRemainingTime = preChangeMinimumRemainingTime THEN
          SELECT MAX(minimumRemainingTime) INTO postChangeMinimumRemainingTime FROM nodes WHERE account=_accountId AND project=_projectId AND parent=_nodeId;
        ELSEIF NOT currentIsParallel THEN
          SET postChangeMinimumRemainingTime = currentMinimumRemainingTime + postChangeMinimumRemainingTime-preChangeMinimumRemainingTime;
        ELSE #nochange to minimum time to make
          SET postChangeMinimumRemainingTime=currentMinimumRemainingTime;
        END IF;
        SET preChangeMinimumRemainingTime=currentMinimumRemainingTime;
      ELSE
        SELECT parent INTO nextNode FROM nodes WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      END IF;

      #do the actual update, needs a bunch of bool logic to work out +/- sign usage, 8 cases for all combinations, but the node is updated in a single update statement :)
      IF totalRemainingTimeChangeIsPositive AND totalLoggedTimeChangeIsPositive AND descendantCountChangeIsPositive THEN
        UPDATE nodes SET totalRemainingTime=totalRemainingTime+totalRemainingTimeChange, totalLoggedTime=totalLoggedTime+totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount+descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      ELSEIF totalRemainingTimeChangeIsPositive AND totalLoggedTimeChangeIsPositive AND NOT descendantCountChangeIsPositive THEN
        UPDATE nodes SET totalRemainingTime=totalRemainingTime+totalRemainingTimeChange, totalLoggedTime=totalLoggedTime+totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount-descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      ELSEIF totalRemainingTimeChangeIsPositive AND NOT totalLoggedTimeChangeIsPositive AND descendantCountChangeIsPositive THEN
        UPDATE nodes SET totalRemainingTime=totalRemainingTime+totalRemainingTimeChange, totalLoggedTime=totalLoggedTime-totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount+descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      ELSEIF totalRemainingTimeChangeIsPositive AND NOT totalLoggedTimeChangeIsPositive AND NOT descendantCountChangeIsPositive THEN
        UPDATE nodes SET totalRemainingTime=totalRemainingTime+totalRemainingTimeChange, totalLoggedTime=totalLoggedTime-totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount-descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      ELSEIF NOT totalRemainingTimeChangeIsPositive AND totalLoggedTimeChangeIsPositive AND descendantCountChangeIsPositive THEN
        UPDATE nodes SET totalRemainingTime=totalRemainingTime-totalRemainingTimeChange, totalLoggedTime=totalLoggedTime+totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount+descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      ELSEIF NOT totalRemainingTimeChangeIsPositive AND totalLoggedTimeChangeIsPositive AND NOT descendantCountChangeIsPositive THEN
        UPDATE nodes SET totalRemainingTime=totalRemainingTime-totalRemainingTimeChange, totalLoggedTime=totalLoggedTime+totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount-descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      ELSEIF NOT totalRemainingTimeChangeIsPositive AND NOT totalLoggedTimeChangeIsPositive AND descendantCountChangeIsPositive THEN
        UPDATE nodes SET totalRemainingTime=totalRemainingTime-totalRemainingTimeChange, totalLoggedTime=totalLoggedTime-totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount+descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      ELSEIF NOT totalRemainingTimeChangeIsPositive AND NOT totalLoggedTimeChangeIsPositive AND NOT descendantCountChangeIsPositive THEN
        UPDATE nodes SET totalRemainingTime=totalRemainingTime-totalRemainingTimeChange, totalLoggedTime=totalLoggedTime-totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount-descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_nodeId;
      END IF;

      SET _nodeId=nextNode;

    END WHILE;
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