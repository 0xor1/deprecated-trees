package account

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

func (s *sqlStore) getAccountRole(shard int, accountId, memberId Id) *AccountRole {
	return GetAccountRole(s.shards[shard], accountId, memberId)
}

func (s *sqlStore) setPublicProjectsEnabled(shard int, accountId Id, publicProjectsEnabled bool) {
	if _, err := s.shards[shard].Exec(`UPDATE accounts SET publicProjectsEnabled=? WHERE id=?`, publicProjectsEnabled, []byte(accountId)); err != nil {
		panic(err)
	}
	if !publicProjectsEnabled {
		if _, err := s.shards[shard].Exec(`UPDATE projects SET isPublic=false WHERE account=?`, []byte(accountId)); err != nil {
			panic(err)
		}
	}
}

func (s *sqlStore) getPublicProjectsEnabled(shard int, accountId Id) bool {
	return GetPublicProjectsEnabled(s.shards[shard], accountId)
}

func (s *sqlStore) setMemberRole(shard int, accountId, memberId Id, role AccountRole) {
	if _, err := s.shards[shard].Exec(`UPDATE accountMembers SET role=? WHERE account=? AND id=?`, role, []byte(accountId), []byte(memberId)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getMember(shard int, accountId, memberId Id) *AccountMember {
	row := s.shards[shard].QueryRow(`SELECT id, name, isActive, role FROM accountMembers WHERE account=? AND id=?`, []byte(accountId), []byte(memberId))
	res := AccountMember{}
	if err := row.Scan(&res.Id, &res.Name, &res.IsActive, &res.Role); err != nil {
		panic(err)
	}
	return &res
}

func (s *sqlStore) getMembers(shard int, accountId Id, role *AccountRole, nameContains *string, offset, limit int) ([]*AccountMember, int) {
	query := bytes.NewBufferString(`SELECT %s FROM accountMembers WHERE account=? AND isActive=true`)
	columns := ` id, name, isActive, role `
	args := make([]interface{}, 0, 5)
	args = append(args, []byte(accountId))
	if role != nil {
		query.WriteString(` AND role=?`)
		args = append(args, role)
	}
	if nameContains != nil {
		query.WriteString(` AND name LIKE ?`)
		strVal := strings.Trim(*nameContains, " ")
		args = append(args, fmt.Sprintf("%%%s%%", strVal))
	}
	count := 0
	row := s.shards[shard].QueryRow(fmt.Sprintf(query.String(), ` COUNT(*) `), args...)
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
	if count == 0 {
		return nil, count
	}
	query.WriteString(` ORDER BY role ASC, name ASC LIMIT ? OFFSET ?`)
	args = append(args, limit, offset)
	rows, err := s.shards[shard].Query(fmt.Sprintf(query.String(), columns), args...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		panic(err)
	}
	res := make([]*AccountMember, 0, limit)
	for rows.Next() {
		mem := AccountMember{}
		if err := rows.Scan(&mem.Id, &mem.Name, &mem.IsActive, &mem.Role); err != nil {
			panic(err)
		}
		res = append(res, &mem)
	}
	return res, count
}

func (s *sqlStore) getActivities(shard int, accountId Id, item *Id, member *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity {
	if occurredAfterUnixMillis != nil && occurredBeforeUnixMillis != nil {
		panic(InvalidArgumentsErr)
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, itemName, action, newValue FROM accountActivities WHERE account=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, []byte(accountId))
	if item != nil {
		query.WriteString(` AND item=?`)
		args = append(args, []byte(*item))
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, []byte(*member))
	}
	if occurredAfterUnixMillis != nil {
		query.WriteString(` AND occurredOn>? ORDER BY occurredOn ASC`)
		args = append(args, occurredAfterUnixMillis)
	}
	if occurredBeforeUnixMillis != nil {
		query.WriteString(` AND occurredOn<? ORDER BY occurredOn DESC`)
		args = append(args, occurredBeforeUnixMillis)
	}
	if occurredAfterUnixMillis == nil && occurredBeforeUnixMillis == nil {
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
		if err := rows.Scan(&unixMilli, &act.Item, &act.Member, &act.ItemType, &act.ItemName, &act.Action, &act.NewValue); err != nil {
			panic(err)
		}
		act.OccurredOn = time.Unix(unixMilli/1000, (unixMilli%1000)*1000000).UTC()
		res = append(res, &act)
	}
	return res
}

func (s *sqlStore) logActivity(shard int, accountId Id, member, item Id, itemType, action string, newValue string) {
	LogAccountActivity(s.shards[shard], accountId, member, item, itemType, action, &newValue)
}
