set -e

COOKIE="login.cookie"

echo "# login to the server"
curl http://localhost:8080/login -XPOST -d "email=user&password=password" -c $COOKIE  
echo ""

echo "# get login status"
curl http://localhost:8080/login -b $COOKIE
echo ""

echo "# access websocket path with http protocol. (it will fail to connect websocke server.)"
curl http://localhost:8080/chat/ws -b $COOKIE
echo ""

# finally remove cookie file
rm $COOKIE
