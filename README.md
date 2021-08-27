# Guid
`guid` is a disributed uique id generation framework written in Go (Golang). `guid` supports two strategies to generate uniqu id: `segment` and `snowflake`

# Required
* go1.16+


# Get Start

```shell
go get github.com/ziidu/guid
```

## snowflake 
> get a id by snowflake method. please see `example/snowflakeuid_example.go`

## segment
> you need to create a table, an insert some data. and see `example/segmentuid_example.go`. you need update `connectURL` in `example/segmentuid_example.go`

```sql
DROP TABLE IF EXISTS `guid`;
CREATE TABLE `guid` (
    -- business field
    `biz_tag` varchar(128) NOT NULL DEFAULT '',
    -- max id for current segment
    `max_id` bigint(20) NOT NULL DEFAULT '0',
    -- max_id increase by step
    `step` int(11) NOT NULL,
    `description` varchar(256) NOT NULL  DEFAULT '',
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`biz_tag`)
) ENGINE=InnoDB;

insert into leaf_alloc 
(biz_tag, max_id, step, description) 
VALUES ('order', 10000, 10000, 'order tag')
```

