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

DROP USER IF EXISTS 'tc_cd_pwds'@'%';
CREATE USER 'tc_cd_pwds'@'%' IDENTIFIED BY 'T@sk-C3n-T3r-Pwd'; 
GRANT SELECT ON pwds.pwds TO 'tc_cd_pwds'@'%'; 
GRANT INSERT ON pwds.pwds TO 'tc_cd_pwds'@'%';
GRANT UPDATE ON pwds.pwds TO 'tc_cd_pwds'@'%';
GRANT DELETE ON pwds.pwds TO 'tc_cd_pwds'@'%';