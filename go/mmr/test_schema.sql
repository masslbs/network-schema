-- SPDX-FileCopyrightText: 2025 Mass Labs
--
-- SPDX-License-Identifier: MIT

CREATE TABLE pgmmr_nodes (
    tree_id BIGINT NOT NULL,
    node_id BIGINT NOT NULL,
    data BYTEA NOT NULL,
    PRIMARY KEY (tree_id, node_id)
);
CREATE UNIQUE INDEX pgmmr_nodes_primary ON pgmmr_nodes(tree_id, node_id);

CREATE TABLE pgmmr_values (
    tree_id BIGINT NOT NULL,
    leaf_idx BIGINT NOT NULL,
    data BYTEA NOT NULL,
    PRIMARY KEY (tree_id, leaf_idx),
    FOREIGN KEY (tree_id, leaf_idx) REFERENCES pgmmr_nodes (tree_id, node_id)
);
CREATE UNIQUE INDEX pgmmr_values_primary ON pgmmr_values(tree_id, leaf_idx);
