package org

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bytes"
	"fmt"
	"github.com/0xor1/isql"
	"strings"
	"time"
)

func newSqlStore(shards map[int]isql.ReplicaSet) store {
	if len(shards) == 0 {
		panic(InvalidArgumentsErr)
	}
	return &sqlStore{
		shards: shards,
	}
}

type sqlStore struct {
	shards map[int]isql.ReplicaSet
}

func (s *sqlStore) setPublicProjectsEnabled(shard int, orgId Id, publicProjectsEnabled bool) {
	if _, err := s.shards[shard].Exec(`UPDATE orgs SET publicProjectsEnabled=? WHERE id=?`, publicProjectsEnabled, []byte(orgId)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getPublicProjectsEnabled(shard int, orgId Id) bool {
	row := s.shards[shard].QueryRow(`SELECT publicProjectsEnabled FROM orgs WHERE id=?`, []byte(orgId))
	res := false
	if err := row.Scan(&res); err != nil {
		panic(err)
	}
	return res
}

func (s *sqlStore) setUserRole(shard int, orgId, userId Id, role OrgRole) {
	if _, err := s.shards[shard].Exec(`UPDATE orgMembers SET role=? WHERE org=? AND id=?`, role, []byte(orgId), []byte(userId)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getMember(shard int, orgId, memberId Id) *Member {
	row := s.shards[shard].QueryRow(`SELECT id, name, totalRemainingTime, totalLoggedTime, isActive, role FROM orgMembers WHERE org=? AND id=?`, []byte(orgId), []byte(memberId))
	res := Member{}
	if err := row.Scan(&res.Id, &res.Name, &res.TotalRemainingTime, &res.TotalLoggedTime, &res.IsActive, &res.Role); err != nil {
		panic(err)
	}
	return &res
}

func (s *sqlStore) getMembers(shard int, orgId Id, role *OrgRole, nameContains *string, offset, limit int) ([]*Member, int) {
	countQuery := bytes.NewBufferString(`SELECT COUNT(*) FROM orgMembers WHERE org=? AND isActive=true`)
	query := bytes.NewBufferString(`SELECT id, name, totalRemainingTime, totalLoggedTime, isActive, role FROM orgMembers WHERE org=? AND isActive=true`)
	args := make([]interface{}, 0, 3)
	args = append(args, []byte(orgId))
	if role != nil {
		countQuery.WriteString(` AND role=?`)
		query.WriteString(` AND role=?`)
		args = append(args, role)
	}
	if nameContains != nil {
		countQuery.WriteString(` AND name LIKE ?`)
		query.WriteString(` AND name LIKE ?`)
		strVal := strings.Trim(*nameContains, " ")
		args = append(args, fmt.Sprintf("%%%s%%", strVal))
	}
	count := 0
	row := s.shards[shard].QueryRow(countQuery.String(), args...)
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
	query.WriteString(` ORDER BY role ASC, name ASC LIMIT ? OFFSET ?`)
	args = append(args, limit, offset)
	rows, err := s.shards[shard].Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		panic(err)
	}
	res := make([]*Member, 0, limit)
	for rows.Next() {
		mem := Member{}
		if err := rows.Scan(&mem.Id, &mem.Name, &mem.TotalRemainingTime, &mem.TotalLoggedTime, &mem.IsActive, &mem.Role); err != nil {
			panic(err)
		}
		res = append(res, &mem)
	}
	return res, count
}

func (s *sqlStore) getActivities(shard int, orgId Id, item *Id, member *Id, occurredAfter *time.Time, occurredBefore *time.Time, limit int) []*Activity {
	if occurredBefore != nil && occurredAfter != nil {
		panic(InvalidArgumentsErr)
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, itemName, action FROM orgActivities WHERE org=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, []byte(orgId))
	if item != nil {
		query.WriteString(` AND item=?`)
		args = append(args, []byte(*item))
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, []byte(*member))
	}
	if occurredAfter != nil {
		query.WriteString(` AND occurredOn>? ORDER BY occurredOn ASC`)
		args = append(args, occurredAfter.UnixNano()/1000000)
	}
	if occurredBefore != nil {
		query.WriteString(` AND occurredOn<? ORDER BY occurredOn DESC`)
		args = append(args, occurredBefore.UnixNano()/1000000)
	}
	if occurredBefore == nil && occurredAfter == nil {
		query.WriteString(` ORDER BY occurredOn DESC`)
	}
	query.WriteString(` LIMIT ?`)
	args = append(args, limit)
	rows, err := s.shards[shard].Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		panic(err)
	}

	res := make([]*Activity, 0, limit)
	for rows.Next() {
		act := Activity{}
		unixMilli := int64(0)
		if err := rows.Scan(&unixMilli, &act.Item, &act.Member, &act.ItemType, &act.ItemName, &act.Action); err != nil {
			panic(err)
		}
		act.OccurredOn = time.Unix(unixMilli/1000, (unixMilli%1000)*1000000).UTC()
		res = append(res, &act)
	}
	return res
}
