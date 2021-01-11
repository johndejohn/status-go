CREATE TABLE communities_requests_to_join  (
  request VARCHAR NOT NULL,
  clock INT NOT NULL,
  ens_name VARCHAR NOT NULL DEFAULT "",
  chat_id VARCHAR NOT NULL DEFAULT "",
  community_id BLOB NOT NULL,
  state INT NOT NULL DEFAULT 0,
  PRIMARY KEY (request, clock)
);
