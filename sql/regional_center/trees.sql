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
    role TINYINT UNSIGNED NOT NULL, #0 owner, 1 admin, 2 standard member
    PRIMARY KEY (org, isActive, role, name),
    UNIQUE INDEX (org, id)
);

DROP TABLE IF EXISTS projects;
CREATE TABLE projects(
	org BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
	name VARCHAR(250) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, id)
);

DROP TABLE IF EXISTS projectMembers;
CREATE TABLE projectMembers(
	org BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    role TINYINT UNSIGNED NOT NULL, #0 owner, 1 admin, 2 standard member
    PRIMARY KEY (org, role, id),
    UNIQUE INDEX (org, id)
);

DROP PROCEDURE IF EXISTS registerOrgAccount;
DELIMITER $$
CREATE PROCEDURE registerOrgAccount(_id BINARY(16), _ownerId BINARY(16), _ownerName VARCHAR(50))
BEGIN
	INSERT INTO orgs (id) VALUES (_id);
    INSERT INTO orgMembers (org, id, name, totalRemainingTime, totalLoggedTime, isActive, isDeleted, role) VALUES (_id, _ownerId, _ownerName, 0, 0, true, false, 0);
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