DROP DATABASE IF EXISTS trees;
CREATE DATABASE trees;
USE trees;

DROP TABLE IF EXISTS members;
CREATE TABLE members(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
    name VARCHAR(50) NOT NULL,
    created DATETIME NOT NULL,
    PRIMARY KEY (name),
    UNIQUE INDEX (id)
);

DROP USER IF EXISTS 'tc_rc_trees'@'%';
CREATE USER 'tc_rc_trees'@'%' IDENTIFIED BY 'T@sk-C3n-T3r';
GRANT SELECT ON accounts.* TO 'tc_rc_trees'@'%';
GRANT INSERT ON accounts.* TO 'tc_rc_trees'@'%';
GRANT UPDATE ON accounts.* TO 'tc_rc_trees'@'%';
GRANT DELETE ON accounts.* TO 'tc_rc_trees'@'%';