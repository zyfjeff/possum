-- See manifest_blocks.sql for the original more complicated schema.

create table if not exists keys (
    key_id integer primary key,
    -- This is to support whatever the OS can use for paths. It was any for a while to support
    -- migrating to different value file naming schemes, but since this is intended for caches,
    -- maybe it's not worth the risk.
    file_id blob,
    file_offset integer,
    value_length integer,
    -- This is the most (concrete?) representation for the finest time granularity sqlite's internal time functions support.
    last_used integer not null default (cast(unixepoch('subsec')*1e3 as integer)),
    -- Put this last because it's most likely looked up in the index and not needed when looking at the row.
    key blob unique not null,
    -- This is necessary for value renames
    unique (file_id, file_offset)
) strict;

create index if not exists last_used_index on keys (
    last_used,
    key_id
);

-- This is for next_value_offset. Does this duplicate the unique (file_id, file_offset) index on keys?
CREATE INDEX if not exists file_id_then_offset on keys (file_id, file_offset);
-- This is for last_end_offset
CREATE INDEX if not exists file_id_then_end_offset on keys (file_id, file_offset+value_length);