docker run \
    --restart always \
    --publish=7474:7474 --publish=7687:7687 \
    --volume=${HOME}/data/neo4j:${HOME}/data/neo4j/databasedata \
    neo4j:5.26.0

