DROP DATABASE IF EXISTS pwds;
CREATE DATABASE pwds;
USE pwds;

DROP TABLE IF EXISTS pwds;
CREATE TABLE pwds(
	id BINARY(16) NOT NULL,
	salt   VARBINARY(256) NOT NULL,
	pwd    VARBINARY(256) NOT NULL,
	n      MEDIUMINT UNSIGNED NOT NULL,
	r      MEDIUMINT UNSIGNED NOT NULL,
	p      MEDIUMINT UNSIGNED NOT NULL,
	keyLen MEDIUMINT UNSIGNED NOT NULL,
    PRIMARY KEY (id)
);

DROP USER IF EXISTS 'task_center_central_directory_api'@'%';
CREATE USER 'task_center_central_directory_api'@'%' IDENTIFIED BY 'T@sk-C3n-T3r-Pwd'; 
GRANT SELECT (id, salt, pwd, n, r, p, keyLen) ON pwds.pwds TO 'task_center_central_directory_api'@'%';