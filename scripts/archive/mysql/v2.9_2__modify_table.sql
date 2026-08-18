SET NAMES utf8mb4;

drop index idx_tag_node on ke_graph_tag_node;

create index idx_tag_node
    on ke_graph_tag_node (tag_id, name);