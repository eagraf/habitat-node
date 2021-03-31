# How to use Filesystem HTTP API:

The filesystem api runs on port 6000. It handles all GET requests of the form /api/fs and works by modifying the request and forwarding it to the IPFS HTTP API. The commands are as follows:


## /api/fs/ls:
curl -s -X GET 'http://127.0.0.1:6000/api/fs/ls?path=<community_id:filename>'

## /api/fs/write:
curl -s -X GET 'http://127.0.0.1:6000/api/fs/write?path=<community_id:filename>&file=<path/to/file/locally>'

## /api/fs/pin:
curl -s -X GET 'http://127.0.0.1:6000/api/fs/pin?path=<community_id:filename>&action=<check | pin | unpin>'

## /api/fs/remove:
curl -s -X GET 'http://127.0.0.1:6000/api/fs/remove?path=<community_id:filename>'

## /api/fs/cat:
curl -s -X GET 'http://127.0.0.1:6000/api/fs/cat?path=<community_id:filename>'

## /api/fs/move:
curl -s -X 'http://127.0.0.1:6000/api/fs/move?old=<community_id:filename>&new=<community_id:filename>'

* note that both community ids must be matching in order for the request to be accepted

## /api/fs/copy:
curl -s -X 'http://127.0.0.1:6000/api/fs/move?old=<community_id:filename>&new=<community_id:filename>'

* note that both community ids must be matching in order for the request to be accepted

## /api/fs/mkdir:
curl -s -X GET 'http://127.0.0.1:6000/api/fs/mkdir?path=<community_id:dirname>'