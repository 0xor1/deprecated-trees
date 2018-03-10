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
  extraInfo VARCHAR(1250) NULL,
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
  extraInfo VARCHAR(1250) NULL,
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

DROP TABLE IF EXISTS tasks;
CREATE TABLE tasks(
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
  task BINARY(16) NOT NULL,
  member BINARY(16) NOT NULL,
  loggedOn DATETIME NOT NULL,
  taskName VARCHAR(250) NOT NULL,
  duration BIGINT UNSIGNED NOT NULL,
  note VARCHAR(250) NULL,
  PRIMARY KEY(account, project, task, loggedOn, member),
  UNIQUE INDEX(account, project, member, loggedOn, task),
  UNIQUE INDEX(account, member, loggedOn, project, task)
);

DROP PROCEDURE IF EXISTS registerAccount;
DELIMITER $$
CREATE PROCEDURE registerAccount(_accountId BINARY(16), _myId BINARY(16), _myName VARCHAR(50), _myDisplayName VARCHAR(100))
BEGIN
	INSERT INTO accounts (id, publicProjectsEnabled) VALUES (_accountId, false);
  INSERT INTO accountMembers (account, id, name, displayName, isActive, role) VALUES (_accountId, _myId, _myName, _myDisplayName, true, 0);
  INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _accountId, 'account', 'created', NULL, NULL);
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
  UPDATE tasks SET member=NULL WHERE account=_account AND member=_member;
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
    DELETE FROM tasks WHERE account=_id;
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
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _accountId, 'account', 'setPublicProjectsEnabled', NULL, 'true');
    ELSE
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _accountId, 'account', 'setPublicProjectsEnabled', NULL, 'false');
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
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _memberId, 'member', 'setRole', NULL, CAST(_role as char character set utf8));
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
    INSERT INTO tasks (account, project, id, parent, firstChild, nextSibling, isAbstract, name, description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member) VALUES (_accountId, _id, _id, NULL, NULL, NULL, TRUE, _name, _description, _createdOn, 0, 0, 0, 0, 0, 0, 0, _isParallel, NULL);
    INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _id, 'project', 'created', _name, NULL);
    INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _id, UTC_TIMESTAMP(6), _myId, _id, 'project', 'created', NULL, NULL);
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
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsPublic', projName, 'true');
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsPublic', NULL, 'true');
    ELSE
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsPublic', projName, 'false');
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsPublic', NULL, 'false');
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
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsArchived', projName, 'true');
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsArchived', NULL, 'true');
    ELSE
      INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsArchived', projName, 'false');
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'setIsArchived', NULL, 'false');
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
	DELETE FROM tasks WHERE account=_accountId AND project=_projectId;
	DELETE FROM timeLogs WHERE account=_accountId AND project=_projectId;
  INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, UTC_TIMESTAMP(6), _myId, _projectId, 'project', 'deleted', projName, NULL);
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
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _id, 'member', 'added', NULL, NULL);
      SET changeMade = true;
    END IF;
	ELSEIF NOT projMemberExists THEN #adding new project member, need to check if they are active account member
		START TRANSACTION;
			SELECT name, displayName INTO accMemberName, accMemberDisplayName FROM accountMembers WHERE account=_accountId AND id=_id AND isActive=true LOCK IN SHARE MODE;
			IF accMemberName IS NOT NULL AND accMemberName <> '' THEN #if active account member then add them to the project
				INSERT INTO projectMembers (account, project, id, name, displayName, isActive, totalRemainingTime, totalLoggedTime, role) VALUES (_accountId, _projectId, _id, accMemberName, accMemberDisplayName, true, 0, 0, _role);
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _id, 'member', 'added', NULL, NULL);
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
  DECLARE projectExists BOOL DEFAULT FALSE;
  DECLARE projMemberExists BOOL DEFAULT FALSE;
  DECLARE changeMade BOOL DEFAULT FALSE;

  START TRANSACTION;
    SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
    IF projectExists THEN
      SELECT COUNT(*)=1 INTO projMemberExists FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_id AND isActive=true FOR UPDATE;
      IF projMemberExists THEN
        UPDATE tasks SET member=NULL WHERE account=_accountId AND project=_projectId AND member=_id;
        UPDATE projectMembers SET totalRemainingTime=0 WHERE account=_accountId AND project=_projectId AND id=_id;
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _id, 'member', 'removed', NULL, NULL);
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
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _memberId, 'member', 'setRole', NULL, CAST(_role as char character set utf8));
    END IF;
    COMMIT;
    SELECT memberExists;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS createTask;
DELIMITER $$
CREATE PROCEDURE createTask(_accountId BINARY(16), _projectId BINARY(16), _parentId BINARY(16), _myId BINARY(16), _previousSiblingId BINARY(16), _taskId BINARY(16), _isAbstract BOOL, _name VARCHAR(250), _description VARCHAR(1250), _createdOn DATETIME, _totalRemainingTime BIGINT UNSIGNED, _isParallel BOOL, _memberId BINARY(16))
BEGIN
	DECLARE projectExists BOOL DEFAULT FALSE;
  DECLARE parentExists BOOL DEFAULT FALSE;
	DECLARE previousSiblingExists BOOL DEFAULT FALSE;
	DECLARE nextSiblingToUse BINARY(16) DEFAULT NULL;
  DECLARE memberExistsAndIsActive BOOL DEFAULT FALSE;
  DECLARE changeMade BOOL DEFAULT FALSE;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  CREATE TEMPORARY TABLE tempUpdatedIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );

  #assume parameters are already validated against themselves in the business logic

  START TRANSACTION;
  #validate parameters against the database, i.e. make sure parent/sibiling/member exist and are active and have sufficient privileges
	#lock the project and lock the project member (member needs "for update" lock if remaining time is greater than zero, otherwise only "lock in share mode" is needed)
	SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
  IF _parentId=_projectId THEN
    SET parentExists=projectExists;
  ELSE
    SELECT COUNT(*)=1 INTO parentExists FROM tasks WHERE account=_accountId AND project=_projectId AND id=_parentId AND isAbstract=TRUE;
  END IF;
  IF _previousSiblingId IS NULL THEN
    SELECT firstChild INTO nextSiblingToUse FROM tasks WHERE account=_accountId AND project=_projectId AND id=_parentId AND isAbstract=TRUE;
    SET previousSiblingExists=TRUE;
	ELSE
    SELECT COUNT(*)=1, nextSibling INTO previousSiblingExists, nextSiblingToUse FROM tasks WHERE account=_accountId AND project=_projectId AND parent=_parentId AND id=_previousSiblingId;
  END IF;
  IF _memberId IS NOT NULL AND _totalRemainingTime <> 0 THEN
    SELECT COUNT(*)=1 INTO memberExistsAndIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE AND role < 2 FOR UPDATE; #less than 2 means 0->projectAdmin or 1->projectWriter
  ELSEIF _memberId IS NOT NULL AND _totalRemainingTime=0 THEN
    SELECT COUNT(*)=1 INTO memberExistsAndIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE AND role < 2 LOCK IN SHARE MODE;
  ELSE
    SET memberExistsAndIsActive=TRUE;
  END IF;
  IF _memberId IS NOT NULL AND _isAbstract THEN
    SET memberExistsAndIsActive=FALSE; #set this to false to fail the validation check so we dont create an invalid abstract task with an assigned member
  END IF;
  IF projectExists AND parentExists AND previousSiblingExists AND memberExistsAndIsActive THEN
    SET changeMade=TRUE;
		#write the task row
    if NOT _isAbstract THEN
      SET _isParallel=FALSE;
    END IF;
    INSERT INTO tasks (account,	project, id, parent, firstChild, nextSibling, isAbstract, name,	description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member) VALUES (_accountId,	_projectId, _taskId, _parentId, NULL, nextSiblingToUse, _isAbstract, _name, _description, _createdOn, _totalRemainingTime, 0, _totalRemainingTime, 0, 0, 0, 0, _isParallel, _memberId);
    INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'created', _name, NULL);
    #update siblings and parent firstChild value if required
    IF _previousSiblingId IS NULL THEN #update parents firstChild
      UPDATE tasks SET firstChild=_taskId WHERE account=_accountId AND project=_projectId AND id=_parentId;
      INSERT INTO tempUpdatedIds VALUES (_parentId) ON DUPLICATE KEY UPDATE id=id;
    ELSE #update previousSiblings nextSibling value
      UPDATE tasks SET nextSibling=_taskId WHERE account=_accountId AND project=_projectId AND id=_previousSiblingId;
      INSERT INTO tempUpdatedIds VALUES (_previousSiblingId) ON DUPLICATE KEY UPDATE id=id;
    END IF;
    #update member if needed
    IF _memberId IS NOT NULL AND _totalRemainingTime <> 0 THEN
      UPDATE projectMembers SET totalRemainingTime=totalRemainingTime + _totalRemainingTime WHERE account=_accountId AND project=_projectId AND id=_memberId;
    END IF;

    CALL _setAncestralChainAggregateValuesFromTask(_accountId, _projectId, _parentId);
  END IF;
  SELECT * FROM tempUpdatedIds;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setTaskName;
DELIMITER $$
CREATE PROCEDURE setTaskName(_accountId BINARY(16), _projectId BINARY(16), _taskId BINARY(16), _myId BINARY(16), _name VARCHAR(250))
  BEGIN
    DECLARE oldName VARCHAR(250) DEFAULT '';
    SELECT name INTO oldName FROM tasks WHERE account=_accountId AND project=_projectId AND id=_taskId;
    UPDATE tasks SET name=_name WHERE account=_accountId AND project=_projectId AND id=_taskId;
    IF _projectId=_taskId THEN
      UPDATE projects SET name=_name WHERE account=_accountId AND id=_taskId;
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'project', 'setName', NULL, oldName);
      UPDATE accountActivities SET itemName=_name WHERE account=_accountId AND item=_taskId;
    ELSE
      UPDATE timeLogs SET taskName=_name WHERE account=_accountId AND project=_projectId AND task=_taskId;
      INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'setName', NULL, oldName);
      UPDATE projectActivities SET itemName=_name WHERE account=_accountId AND project=_projectId AND item=_taskId;
    END IF;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setTaskDescription;
DELIMITER $$
CREATE PROCEDURE setTaskDescription(_accountId BINARY(16), _projectId BINARY(16), _taskId BINARY(16), _myId BINARY(16), _description VARCHAR(250))
  BEGIN
    UPDATE tasks SET description=_description WHERE account=_accountId AND project=_projectId AND id=_taskId;
    INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'setDescription', NULL, _description);
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setTaskIsParallel;
DELIMITER $$
CREATE PROCEDURE setTaskIsParallel(_accountId BINARY(16), _projectId BINARY(16), _taskId BINARY(16), _myId BINARY(16), _isParallel BOOL)
BEGIN
	DECLARE projectExists BOOL DEFAULT FALSE;
  DECLARE taskExists BOOL DEFAULT FALSE;
  DECLARE nextTask BINARY(16) DEFAULT NULL;
  DECLARE currentIsParallel BOOL DEFAULT FALSE;
  DECLARE currentMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE changeMade BOOL DEFAULT FALSE;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  CREATE TEMPORARY TABLE tempUpdatedIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );
  START TRANSACTION;
  SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE; #set project lock to ensure data integrity
  IF projectExists THEN
    SELECT COUNT(*)=1, parent, isParallel, minimumRemainingTime INTO taskExists, nextTask, currentIsParallel, currentMinimumRemainingTime FROM tasks WHERE account=_accountId AND project=_projectId AND id=_taskId;
    IF taskExists AND _isParallel <> currentIsParallel THEN #make sure we are making a change otherwise, no need to update anything
      SET changeMade = TRUE;
      IF _isParallel THEN
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'setIsParallel', NULL, 'true');
      ELSE
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'setIsParallel', NULL, 'false');
      END IF;
      UPDATE tasks SET isParallel=_isParallel WHERE account=_accountId AND project=_projectId AND id=_taskId;
      IF currentMinimumRemainingTime <> 0 THEN
        CALL _setAncestralChainAggregateValuesFromTask(_accountId, _projectId, _taskId);
      END IF;
    END IF;
  END IF;
  SELECT * FROM tempUpdatedIds;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS setTaskMember;
DELIMITER $$
CREATE PROCEDURE setTaskMember(_accountId BINARY(16), _projectId BINARY(16), _taskId BINARY(16), _myId BINARY(16), _memberId BINARY(16))
BEGIN
  DECLARE projectExists BOOL DEFAULT FALSE;
  DECLARE memberExistsAndIsActive BOOL DEFAULT TRUE;
  DECLARE taskExistsAndIsConcrete BOOL DEFAULT FALSE;
  DECLARE changeMade BOOL DEFAULT FALSE;
  DECLARE taskTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE existingMemberTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE existingMember BINARY(16) DEFAULT NULL;
  DECLARE existingMemberExists BOOL DEFAULT FALSE; #seems pointless but it's justa var to stick a value in when locking the row
  START TRANSACTION;
  SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE; #set project lock to ensure data integrity
  IF projectExists THEN
    SELECT COUNT(*)=1, totalRemainingTime, member INTO taskExistsAndIsConcrete, taskTotalRemainingTime, existingMember FROM tasks WHERE account=_accountId AND project=_projectId AND id=_taskId AND isAbstract=FALSE;
    IF (existingMember <> _memberId OR (existingMember IS NOT NULL AND _memberId IS NULL) OR (existingMember IS NULL AND _memberId IS NOT NULL)) AND taskExistsAndIsConcrete THEN
      IF taskTotalRemainingTime > 0 THEN
        IF _memberId IS NOT NULL THEN
          SELECT COUNT(*)=1 INTO memberExistsAndIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE AND role < 2 FOR UPDATE; #less than 2 means 0->projectAdmin or 1->projectWriter
          IF memberExistsAndIsActive THEN
            UPDATE projectMembers SET totalRemainingTime=totalRemainingTime+taskTotalRemainingTime WHERE account=_accountId AND project=_projectId AND id=_memberId;
          END IF;
        END IF;
        IF existingMember IS NOT NULL THEN
          SELECT COUNT(*)=1, totalRemainingTime INTO existingMemberExists, existingMemberTotalRemainingTime FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=existingMember FOR UPDATE; #less than 2 means 0->projectAdmin or 1->projectWriter
          IF existingMemberTotalRemainingTime >= taskTotalRemainingTime THEN
            UPDATE projectMembers SET totalRemainingTime=totalRemainingTime-taskTotalRemainingTime WHERE account=_accountId AND project=_projectId AND id=existingMember;
          END IF;
        END IF;
      ELSE
        IF _memberId IS NOT NULL THEN
          SELECT COUNT(*)=1 INTO memberExistsAndIsActive FROM projectMembers WHERE account=_accountId AND project=_projectId AND id=_memberId AND isActive=TRUE AND role < 2 LOCK IN SHARE MODE; #less than 2 means 0->projectAdmin or 1->projectWriter
        END IF;
      END IF;
      IF memberExistsAndIsActive THEN
        UPDATE tasks SET member=_memberId WHERE account=_accountId AND project=_projectId AND id=_taskId;
        SET changeMade = TRUE;
        IF _memberId IS NOT NULL THEN
          INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'setMember', NULL, LOWER(HEX(_memberId)));
        ELSE
          INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'setMember', NULL, NULL);
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
CREATE PROCEDURE setRemainingTimeAndOrLogTime(_accountId BINARY(16), _projectId BINARY(16), _taskId BINARY(16), _myId bINARY(16), _timeRemaining BIGINT UNSIGNED, _loggedOn DATETIME, _duration BIGINT UNSIGNED, _note VARCHAR(250))
BEGIN
  DECLARE projectExists BOOL DEFAULT FALSE;
  DECLARE taskExists BOOL DEFAULT FALSE;
  DECLARE memberExists BOOL DEFAULT FALSE;
  DECLARE taskName VARCHAR(250) DEFAULT '';
  DECLARE existingMemberExists BOOL DEFAULT FALSE;
  DECLARE existingMember BINARY(16) DEFAULT NULL;
  DECLARE nextTask BINARY(16) DEFAULT NULL;
  DECLARE originalMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE changeMade BOOL DEFAULT FALSE;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  CREATE TEMPORARY TABLE tempUpdatedIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );
  START TRANSACTION;
  SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
  IF projectExists THEN
    SELECT COUNT(*)=1, name, member, parent, totalRemainingTime INTO taskExists, taskName, existingMember, nextTask, originalMinimumRemainingTime FROM tasks WHERE account=_accountId AND project=_projectId AND id=_taskId AND isAbstract=FALSE;
    IF _timeRemaining IS NULL THEN
      SET _timeRemaining = originalMinimumRemainingTime;
    END IF;
    IF _duration IS NULL THEN
      SET _duration=0;
    END IF;
    IF taskExists AND (originalMinimumRemainingTime <> _timeRemaining OR _duration > 0) THEN
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
          INSERT INTO timeLogs (account, project, task, member, loggedOn, taskName, duration, note) VALUES (_accountId, _projectId, _taskId, _myId, _loggedOn, taskName, _duration, _note);
          UPDATE projectMembers SET totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_myId;
          INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'loggedTime', taskName, _note);
        END IF;
      END IF;

      UPDATE tasks SET totalRemainingTime=_timeRemaining, minimumRemainingTime=_timeRemaining, totalLoggedTime=totalLoggedTime+_duration WHERE account=_accountId AND project=_projectId AND id=_taskId;
      INSERT INTO tempUpdatedIds VALUES (_taskId) ON DUPLICATE KEY UPDATE id=id;
      IF _timeRemaining <> originalMinimumRemainingTime THEN
        INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, ADDTIME(UTC_TIMESTAMP(6), '0:0:0.000001'), _myId, _taskId, 'task', 'setRemainingTime', taskName, CAST(_timeRemaining as char character set utf8));
      END IF;
      CALL _setAncestralChainAggregateValuesFromTask(_accountId, _projectId, nextTask);
    END IF;
  END IF;
  SELECT * FROM tempUpdatedIds;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS moveTask;
DELIMITER $$
CREATE PROCEDURE moveTask(_accountId BINARY(16), _projectId BINARY(16), _taskId BINARY(16), _newParentId BINARY(16), _myId BINARY(16), _newPreviousSiblingId BINARY(16))
CONTAINS SQL moveTask:
BEGIN
  DECLARE projectExists BOOL DEFAULT FALSE;
  DECLARE taskExists BOOL DEFAULT FALSE;
  DECLARE taskName VARCHAR(250) DEFAULT '';
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
  DECLARE idVariable BINARY(16) DEFAULT _newParentId;
  DECLARE changeMade BOOL DEFAULT FALSE;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  CREATE TEMPORARY TABLE tempUpdatedIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );
  START TRANSACTION;
  IF _newParentId IS NOT NULL AND _projectId <> _taskId AND _newParentId <> _taskId THEN #check newParent is not null AND _taskId is not the project or the new parent
    SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
    IF projectExists THEN
      SELECT COUNT(*)=1, parent, nextSibling, name, totalRemainingTime, totalLoggedTime, minimumRemainingTime INTO taskExists, originalParentId, originalNextSiblingId, taskName, originalTotalRemainingTime, originalTotalLoggedTime, originalMinimumRemainingTime FROM tasks WHERE account=_accountId AND project=_projectId AND id=_taskId;
      IF taskExists THEN
        SELECT firstChild INTO originalParentFirstChildId FROM tasks WHERE account=_accountId AND project=_projectId AND id=originalParentId;
        IF _newPreviousSiblingId IS NOT NULL THEN
          SELECT COUNT(*)=1, nextSibling INTO newPreviousSiblingExists, newNextSiblingId FROM tasks WHERE account=_accountId AND project=_projectId AND id=_newPreviousSiblingId AND parent=_newParentId;
        ELSE
          SET newPreviousSiblingExists=TRUE;
        END IF;
        IF newPreviousSiblingExists THEN
          IF _newParentId = originalParentId THEN #most moves will be just moving within a single parent task, this is a very simple and efficient move to make as we dont need to walk up the tree updating child/descendant counts or total/minimum/logged times
            IF _newPreviousSiblingId IS NULL AND originalParentFirstChildId <> _taskId  THEN #moving to firstChild position, and it is an actual change
              SELECT id INTO originalPreviousSiblingId FROM tasks WHERE account=_accountId AND project=_projectId AND nextSibling=_taskId;
              UPDATE tasks SET nextSibling=originalParentFirstChildId WHERE account=_accountId AND project=_projectId AND id=_taskId;
              UPDATE tasks SET nextSibling=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalPreviousSiblingId;
              UPDATE tasks SET firstChild=_taskId WHERE account=_accountId AND project=_projectId AND id=originalParentId;
              INSERT INTO tempUpdatedIds VALUES (_taskId), (originalPreviousSiblingId), (originalParentId) ON DUPLICATE KEY UPDATE id=id;
              SET changeMade = TRUE;
            ELSEIF _newPreviousSiblingId IS NOT NULL AND ((newNextSiblingId IS NULL AND originalNextSiblingId IS NOT NULL) OR (newNextSiblingId IS NOT NULL AND newNextSiblingId <> _taskId AND originalNextSiblingId IS NULL) OR (newNextSiblingId IS NOT NULL AND newNextSiblingId <> _taskId AND originalNextSiblingId IS NOT NULL AND newNextSiblingId <> originalNextSiblingId)) THEN
              IF originalParentFirstChildId = _taskId AND originalNextSiblingId IS NOT NULL THEN #moving the first child task
                UPDATE tasks SET firstChild=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalParentId;
                INSERT INTO tempUpdatedIds VALUES (originalParentId) ON DUPLICATE KEY UPDATE id=id;
              ELSE #moving a none first child task
                SELECT id INTO originalPreviousSiblingId FROM tasks WHERE account=_accountId AND project=_projectId AND nextSibling=_taskId;
                IF originalPreviousSiblingId IS NOT NULL THEN
                  UPDATE tasks SET nextSibling=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalPreviousSiblingId;
                  INSERT INTO tempUpdatedIds VALUES (originalPreviousSiblingId) ON DUPLICATE KEY UPDATE id=id;
                END IF;
              END IF;
              UPDATE tasks SET nextSibling=newNextSiblingId WHERE account=_accountId AND project=_projectId AND id=_taskId;
              UPDATE tasks SET nextSibling=_taskId WHERE account=_accountId AND project=_projectId AND id=_newPreviousSiblingId;
              INSERT INTO tempUpdatedIds VALUES (_taskId), (_newPreviousSiblingId) ON DUPLICATE KEY UPDATE id=id;
              SET changeMade = TRUE;
            END IF;
          ELSE # this is the expensive complex move to make, to simplify we do not work out a shared ancestor task and perform operations up to that task
               # we fully remove the aggregated values all the way to the project task, then add them back in in the new location, this may result in more processing,
               # but should still be efficient and simplify the code logic below.
            SELECT COUNT(*)=1, firstChild INTO newParentExists, newParentFirstChildId FROM tasks WHERE account=_accountId AND project=_projectId AND id=_newParentId;
            IF newParentExists THEN
              #need to validate that the task being moved is not in the new ancestral chain i.e. make sure we're not trying to make it a descendant of itself
              WHILE idVariable IS NOT NULL DO
                IF idVariable = _taskId THEN
                  SELECT changeMade;
                  COMMIT; #invalid call, exit immediately
                  LEAVE moveTask;
                END IF;
                SELECT parent INTO idVariable FROM tasks WHERE account=_accountId AND project=_projectId AND id=idVariable;
              END WHILE;
              #move the task
              #remove from original location
              IF originalParentFirstChildId = _taskId THEN #removing from first child position
                UPDATE tasks SET firstChild=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalParentId;
                INSERT INTO tempUpdatedIds VALUES (originalParentId) ON DUPLICATE KEY UPDATE id=id;
              ELSE #removing from none first child position
                SELECT id INTO originalPreviousSiblingId FROM tasks WHERE account=_accountId AND project=_projectId AND nextSibling=_taskId;
                UPDATE tasks SET nextSibling=originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalPreviousSiblingId;
                INSERT INTO tempUpdatedIds VALUES (originalPreviousSiblingId) ON DUPLICATE KEY UPDATE id=id;
              END IF;
              ##remove parent id so it isn't included in update in setAncestralChainAggregateValuesFromTask call
              UPDATE tasks SET parent=NULL WHERE account=_accountId AND project=_projectId AND id=_taskId;
              INSERT INTO tempUpdatedIds VALUES (_taskId) ON DUPLICATE KEY UPDATE id=id;
              CALL _setAncestralChainAggregateValuesFromTask(_accountId, _projectId, originalParentId);

              #now add in the new position
              IF _newPreviousSiblingId IS NULL THEN #moving to firstChild position
                UPDATE tasks SET firstChild=_taskId WHERE account=_accountId AND project=_projectId AND id=_newParentId;
                UPDATE tasks SET parent=_newParentId, nextSibling=newParentFirstChildId WHERE account=_accountId AND project=_projectId AND id=_taskId;
                INSERT INTO tempUpdatedIds VALUES (_newParentId), (_taskId) ON DUPLICATE KEY UPDATE id=id;
              ELSE #moving to none firstChild position
                UPDATE tasks SET nextSibling=_taskId WHERE account=_accountId AND project=_projectId AND id=_newPreviousSiblingId;
                UPDATE tasks SET parent=_newParentId, nextSibling=newNextSiblingId WHERE account=_accountId AND project=_projectId AND id=_taskId;
                INSERT INTO tempUpdatedIds VALUES (_newPreviousSiblingId), (_taskId) ON DUPLICATE KEY UPDATE id=id;
              END IF;
              CALL _setAncestralChainAggregateValuesFromTask(_accountId, _projectId, _newParentId);
              SET changeMade=TRUE;
            END IF;
          END IF;
        END IF;
      END IF;
    END IF;
  END IF;
  IF changeMade THEN
    INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'moved', taskName, NULL);
  END IF;
  SELECT * FROM tempUpdatedIds;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteTask;
DELIMITER $$
CREATE PROCEDURE deleteTask(_accountId BINARY(16), _projectId BINARY(16), _taskId BINARY(16), _myId BINARY(16))
BEGIN
  DECLARE projectExists BOOL DEFAULT FALSE;
  DECLARE taskExists BOOL DEFAULT FALSE;
  DECLARE originalParentId BINARY(16) DEFAULT NULL;
  DECLARE originalNextSiblingId BINARY(16) DEFAULT NULL;
  DECLARE originalPreviousSiblingId BINARY(16) DEFAULT NULL;
  DECLARE originalDescendantCount BIGINT UNSIGNED DEFAULT 0;
  DECLARE originalTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE originalTotalLoggedTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE originalMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
  DECLARE deleteCount BIGINT UNSIGNED DEFAULT 0;
  DECLARE taskName VARCHAR(250) DEFAULT '';
  DECLARE changeMade BOOL DEFAULT FALSE;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  CREATE TEMPORARY TABLE tempUpdatedIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );
  DROP TEMPORARY TABLE IF EXISTS tempAllIds;
  CREATE TEMPORARY TABLE tempAllIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );
  DROP TEMPORARY TABLE IF EXISTS tempCurrentIds;
  CREATE TEMPORARY TABLE tempCurrentIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );
  DROP TEMPORARY TABLE IF EXISTS tempLatestIds;
  CREATE TEMPORARY TABLE tempLatestIds(
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
  );
  START TRANSACTION;

  IF _projectId <> _taskId THEN
    SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId FOR UPDATE;
    IF projectExists THEN
      SELECT COUNT(*)=1, parent, descendantCount, totalRemainingTime, totalLoggedTime, minimumRemainingTime, nextSibling, name INTO taskExists, originalParentId, originalDescendantCount, originalTotalRemainingTime, originalTotalLoggedTime, originalMinimumRemainingTime, originalNextSiblingId, taskName FROM tasks WHERE account=_accountId AND project=_projectId AND id=_taskId;
      IF taskExists THEN
        SELECT id INTO originalPreviousSiblingId FROM tasks WHERE account=_accountId AND project=_projectId AND nextSibling=_taskId;
        INSERT INTO tempCurrentIds VALUES (_taskId);
        WHILE (SELECT COUNT(*) FROM tempCurrentIds) > 0 DO
          INSERT INTO tempLatestIds SELECT id FROM tasks WHERE account=_accountId AND project=_projectId AND parent IN (SELECT id FROM tempCurrentIds);
          INSERT INTO tempAllIds SELECT id FROM tempCurrentIds;
          TRUNCATE tempCurrentIds;
          INSERT INTO tempCurrentIds SELECT id FROM tempLatestIds;
          TRUNCATE tempLatestIds;
        END WHILE;
        SELECT COUNT(*) INTO deleteCount FROM tempAllIds;
        IF deleteCount = originalDescendantCount + 1 THEN
          IF originalPreviousSiblingId IS NULL THEN
            UPDATE tasks SET firstChild = originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalParentId;
          ELSE
            UPDATE tasks SET nextSibling = originalNextSiblingId WHERE account=_accountId AND project=_projectId AND id=originalPreviousSiblingId;
          END IF;
          DELETE FROM tasks WHERE id IN (SELECT id FROM tempAllIds);
          INSERT INTO tempUpdatedIds SELECT id FROM tempAllIds tmpAll ON DUPLICATE KEY UPDATE id=tmpAll.id;
          CALL _setAncestralChainAggregateValuesFromTask(_accountId, _projectId, originalParentId);
          SET changeMade = TRUE;
        END IF;
      END IF;
    END IF;
  END IF;
  IF changeMade THEN
    INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (_accountId, _projectId, UTC_TIMESTAMP(6), _myId, _taskId, 'task', 'deleted', taskName, CONCAT('{"totalRemainingTime":', CAST(originalTotalRemainingTime as char character set utf8), ',"totalLoggedTime":', CAST(originalTotalLoggedTime as char character set utf8), ',"descendantCount":', CAST(originalDescendantCount as char character set utf8), '}'));
  END IF;

  SELECT * FROM tempUpdatedIds;
  DROP TEMPORARY TABLE IF EXISTS tempUpdatedIds;
  DROP TEMPORARY TABLE IF EXISTS tempAllIds;
  DROP TEMPORARY TABLE IF EXISTS tempCurrentIds;
  DROP TEMPORARY TABLE IF EXISTS tempLatestIds;
  COMMIT;
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS getTasks;
DELIMITER $$
CREATE PROCEDURE getTasks(_accountId BINARY(16), _projectId BINARY(16), _taskIdsStr VARCHAR(16000)) #16000 == 500 uuids
  BEGIN
    DECLARE projectExists BOOL DEFAULT FALSE;
    DECLARE taskIdsStrLen INT DEFAULT LENGTH(_taskIdsStr);
    DECLARE offset INT DEFAULT 0;
    DROP TEMPORARY TABLE IF EXISTS tempIds;
    CREATE TEMPORARY TABLE tempIds(
      id BINARY(16) NOT NULL,
      PRIMARY KEY (id)
    );
    START TRANSACTION;

    SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId LOCK IN SHARE MODE;
    IF projectExists AND taskIdsStrLen > 0 AND taskIdsStrLen % 32 = 0 THEN
      WHILE offset < taskIdsStrLen DO
        INSERT INTO tempIds VALUE (UNHEX(SUBSTRING(_taskIdsStr, offset + 1, 32)));
        SET offset = offset + 32;
      END WHILE;
    END IF;
    SELECT id, isAbstract, name, description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member FROM tasks WHERE account=_accountId AND project=_projectId AND id IN (SELECT id FROM tempIds);
    DROP TEMPORARY TABLE IF EXISTS tempIds;
    COMMIT;
  END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS getChildTasks;
DELIMITER $$
CREATE PROCEDURE getChildTasks(_accountId BINARY(16), _projectId BINARY(16), _parentId BINARY(16), _fromSiblingId BINARY(16), _limit INT)
  BEGIN
    DECLARE projectExists BOOL DEFAULT FALSE;
    DECLARE idVariable BINARY(16) DEFAULT NULL;
    DECLARE idx INT DEFAULT 0;
    DROP TEMPORARY TABLE IF EXISTS tempResult;
    CREATE TEMPORARY TABLE tempResult(
      selectOrder INT NOT NULL,
      nextSibling BINARY(16) NULL,
      id BINARY(16) NOT NULL,
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
      PRIMARY KEY (selectOrder)
    );
    START TRANSACTION;

    SELECT COUNT(*)=1 INTO projectExists FROM projectLocks WHERE account=_accountId AND id=_projectId LOCK IN SHARE MODE;
    IF projectExists THEN
      IF _fromSiblingId IS NOT NULL THEN
        SELECT nextSibling INTO idVariable FROM tasks WHERE account=_accountId AND project=_projectId AND id=_fromSiblingId AND parent=_parentId;
      ELSE
        SELECT firstChild INTO idVariable FROM tasks WHERE account=_accountId AND project=_projectId AND id=_parentId;
      END IF;
      WHILE idVariable IS NOT NULL AND idx < _limit DO
        INSERT INTO tempResult SELECT idx, nextSibling, id, isAbstract, name, description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member FROM tasks WHERE account=_accountId AND project=_projectId AND id=idVariable;
        SELECT nextSibling INTO idVariable FROM tempResult WHERE selectOrder = idx;
        SET idx = idx + 1;
      END WHILE;
    END IF;
    SELECT id, isAbstract, name, description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member FROM tempResult ORDER BY selectOrder ASC;
    DROP TEMPORARY TABLE IF EXISTS tempResult;
    COMMIT;
  END;
$$
DELIMITER ;

#!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!#
#********************************MAGIC PROCEDURE WARNING*********************************#
# THIS PROCEDURE MUST ONLY BE CALLED INTERNALLY BY THE ABOVE STORED PROCEDURES THAT HAVE #
# SET THEIR OWN TRANSACTIONS AND PROJECTID LOCKS AND HAVE VALIDATED ALL INPUT PARAMS.    #                                                                                                  #
#!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!#
DROP PROCEDURE IF EXISTS _setAncestralChainAggregateValuesFromTask;
DELIMITER $$
CREATE PROCEDURE _setAncestralChainAggregateValuesFromTask(_accountId BINARY(16), _projectId BINARY(16), _taskId BINARY(16))
  BEGIN
    DECLARE originalTotalRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE originalTotalLoggedTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE currentMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE originalChildCount BIGINT UNSIGNED DEFAULT 0;
    DECLARE originalDescendantCount BIGINT UNSIGNED DEFAULT 0;
    DECLARE currentIsParallel BOOL DEFAULT FALSE;
    DECLARE nextTask BINARY(16) DEFAULT NULL;
    DECLARE totalRemainingTimeChange BIGINT UNSIGNED DEFAULT 0;
    DECLARE totalLoggedTimeChange BIGINT UNSIGNED DEFAULT 0;
    DECLARE preChangeMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE postChangeMinimumRemainingTime BIGINT UNSIGNED DEFAULT 0;
    DECLARE newChildCount BIGINT UNSIGNED DEFAULT 0;
    DECLARE descendantCountChange BIGINT UNSIGNED DEFAULT 0;
    DECLARE totalRemainingTimeChangeIsPositive BOOL DEFAULT TRUE;
    DECLARE totalLoggedTimeChangeIsPositive BOOL DEFAULT TRUE;
    DECLARE descendantCountChangeIsPositive BOOL DEFAULT TRUE;

    SELECT totalRemainingTime, totalLoggedTime, minimumRemainingTime, childCount, descendantCount, isParallel, parent INTO originalTotalRemainingTime, originalTotalLoggedTime, preChangeMinimumRemainingTime, originalChildCount, originalDescendantCount, currentIsParallel, nextTask FROM tasks WHERE account = _accountId AND project = _projectId AND id = _taskId;
    IF currentIsParallel THEN
      SELECT SUM(totalRemainingTime), SUM(totalLoggedTime), MAX(minimumRemainingTime), COUNT(*), SUM(descendantCount)INTO totalRemainingTimeChange, totalLoggedTimeChange, postChangeMinimumRemainingTime, newChildCount, descendantCountChange FROM tasks WHERE account = _accountId AND project = _projectId AND parent = _taskId;
    ELSE                                                   #this is the only difference#
      SELECT SUM(totalRemainingTime), SUM(totalLoggedTime), SUM(minimumRemainingTime), COUNT(*), SUM(descendantCount)INTO totalRemainingTimeChange, totalLoggedTimeChange, postChangeMinimumRemainingTime, newChildCount, descendantCountChange FROM tasks WHERE account = _accountId AND project = _projectId AND parent = _taskId;
    END IF;
    SET descendantCountChange = descendantCountChange + newChildCount;

    #the first task updated is special, it could have had a new child added or removed from it, so the childCount can be updated, no other ancestor will have the childCount updated
    UPDATE tasks SET totalRemainingTime = totalRemainingTimeChange, totalLoggedTime = totalLoggedTimeChange, minimumRemainingTime = postChangeMinimumRemainingTime, childCount = newChildCount, descendantCount = descendantCountChange WHERE account = _accountId AND project = _projectId AND id = _taskId;
    INSERT INTO tempUpdatedIds VALUES (_taskId) ON DUPLICATE KEY UPDATE id=id;

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

    SET _taskId= nextTask;

    WHILE _taskId IS NOT NULL AND (totalRemainingTimeChange > 0 OR totalLoggedTimeChange > 0 OR preChangeMinimumRemainingTime <> postChangeMinimumRemainingTime OR descendantCountChange > 0) DO
      IF preChangeMinimumRemainingTime <> postChangeMinimumRemainingTime THEN #updating minimumRemainingTime and others
        #get values needed to update current task
        SELECT isParallel, minimumRemainingTime, parent INTO currentIsParallel, currentMinimumRemainingTime, nextTask FROM tasks WHERE account=_accountId AND project=_projectId AND id=_taskId;
        IF currentIsParallel AND currentMinimumRemainingTime < postChangeMinimumRemainingTime THEN
          SET postChangeMinimumRemainingTime = postChangeMinimumRemainingTime; #pointless assignment but this if case is necessary
        ELSEIF currentIsParallel AND currentMinimumRemainingTime = preChangeMinimumRemainingTime THEN
          SELECT MAX(minimumRemainingTime) INTO postChangeMinimumRemainingTime FROM tasks WHERE account=_accountId AND project=_projectId AND parent=_taskId;
        ELSEIF NOT currentIsParallel THEN
          SET postChangeMinimumRemainingTime = currentMinimumRemainingTime + postChangeMinimumRemainingTime-preChangeMinimumRemainingTime;
        ELSE #nochange to minimum time to make
          SET postChangeMinimumRemainingTime=currentMinimumRemainingTime;
        END IF;
        SET preChangeMinimumRemainingTime=currentMinimumRemainingTime;
      ELSE
        SELECT parent INTO nextTask FROM tasks WHERE account=_accountId AND project=_projectId AND id=_taskId;
      END IF;

      #do the actual update, needs a bunch of bool logic to work out +/- sign usage, 8 cases for all combinations, but the task is updated in a single update statement :)
      IF totalRemainingTimeChangeIsPositive AND totalLoggedTimeChangeIsPositive AND descendantCountChangeIsPositive THEN
        UPDATE tasks SET totalRemainingTime=totalRemainingTime+totalRemainingTimeChange, totalLoggedTime=totalLoggedTime+totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount+descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_taskId;
      ELSEIF totalRemainingTimeChangeIsPositive AND totalLoggedTimeChangeIsPositive AND NOT descendantCountChangeIsPositive THEN
        UPDATE tasks SET totalRemainingTime=totalRemainingTime+totalRemainingTimeChange, totalLoggedTime=totalLoggedTime+totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount-descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_taskId;
      ELSEIF totalRemainingTimeChangeIsPositive AND NOT totalLoggedTimeChangeIsPositive AND descendantCountChangeIsPositive THEN
        UPDATE tasks SET totalRemainingTime=totalRemainingTime+totalRemainingTimeChange, totalLoggedTime=totalLoggedTime-totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount+descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_taskId;
      ELSEIF totalRemainingTimeChangeIsPositive AND NOT totalLoggedTimeChangeIsPositive AND NOT descendantCountChangeIsPositive THEN
        UPDATE tasks SET totalRemainingTime=totalRemainingTime+totalRemainingTimeChange, totalLoggedTime=totalLoggedTime-totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount-descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_taskId;
      ELSEIF NOT totalRemainingTimeChangeIsPositive AND totalLoggedTimeChangeIsPositive AND descendantCountChangeIsPositive THEN
        UPDATE tasks SET totalRemainingTime=totalRemainingTime-totalRemainingTimeChange, totalLoggedTime=totalLoggedTime+totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount+descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_taskId;
      ELSEIF NOT totalRemainingTimeChangeIsPositive AND totalLoggedTimeChangeIsPositive AND NOT descendantCountChangeIsPositive THEN
        UPDATE tasks SET totalRemainingTime=totalRemainingTime-totalRemainingTimeChange, totalLoggedTime=totalLoggedTime+totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount-descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_taskId;
      ELSEIF NOT totalRemainingTimeChangeIsPositive AND NOT totalLoggedTimeChangeIsPositive AND descendantCountChangeIsPositive THEN
        UPDATE tasks SET totalRemainingTime=totalRemainingTime-totalRemainingTimeChange, totalLoggedTime=totalLoggedTime-totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount+descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_taskId;
      ELSEIF NOT totalRemainingTimeChangeIsPositive AND NOT totalLoggedTimeChangeIsPositive AND NOT descendantCountChangeIsPositive THEN
        UPDATE tasks SET totalRemainingTime=totalRemainingTime-totalRemainingTimeChange, totalLoggedTime=totalLoggedTime-totalLoggedTimeChange, minimumRemainingTime=postChangeMinimumRemainingTime, descendantCount=descendantCount-descendantCountChange WHERE account=_accountId AND project=_projectId AND id=_taskId;
      END IF;

      INSERT INTO tempUpdatedIds VALUES (_taskId) ON DUPLICATE KEY UPDATE id=id;

      SET _taskId=nextTask;

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

#useful helper query for manual verifying results
#SELECT  t1.name, t2.name AS parent, t3.name AS nextSibling, t4.name AS firstChild, t1.isAbstract, t1.description, t1.totalRemainingTime, t1.totalLoggedTime, t1.minimumRemainingTime, t1.childCount, t1.descendantCount FROM trees.tasks t1 LEFT JOIN trees.tasks t2 ON t1.parent = t2.id LEFT JOIN trees.tasks t3 ON t1.nextSibling = t3.id LEFT JOIN trees.tasks t4 ON t1.firstChild = t4.id ORDER BY t1.name;