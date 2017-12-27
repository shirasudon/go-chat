set -e

COOKIE="login.cookie"

echo "# login to the server"
curl http://localhost:8080/login -XPOST -d "name=user2&password=password" -c $COOKIE  
echo ""

echo "# get login status"
curl http://localhost:8080/login -b $COOKIE
echo ""

echo "# create new room"
curl http://localhost:8080/chat/rooms -XPOST -d '{ "name": "new room"}' -b $COOKIE -H "Content-type: application/json"
echo ""

echo "# get room info"
curl http://localhost:8080/chat/rooms/4 -b $COOKIE
echo ""

echo "# post new message"
curl http://localhost:8080/chat/rooms/4/messages -b $COOKIE \
  -XPOST -d '{ "content": "hello! new room!"}' -H "Content-type: application/json"
echo ""

echo "# get room messages (1)"
curl http://localhost:8080/chat/rooms/4/messages -b $COOKIE
echo ""

echo "# get unread room messages (1)"
curl http://localhost:8080/chat/rooms/4/messages/unread -b $COOKIE
echo ""

echo "# read new message"
curl http://localhost:8080/chat/rooms/4/messages/read -b $COOKIE \
  -XPOST -d '{}' -H "Content-type: application/json"
echo ""

echo "# get room messages (2)"
curl http://localhost:8080/chat/rooms/4/messages -b $COOKIE
echo ""

echo "# get unread room messages (2)"
curl http://localhost:8080/chat/rooms/4/messages/unread -b $COOKIE
echo ""

echo "# access websocket path with http protocol. (it will fail to connect websocke server.)"
curl http://localhost:8080/chat/ws -b $COOKIE
echo ""

# finally remove cookie file
rm $COOKIE
