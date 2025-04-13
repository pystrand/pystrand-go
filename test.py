import socket
import json
import time
import threading
import uuid
from typing import Dict, Set

class TCPBackendClient:
    def __init__(self, host="localhost", port=8081):
        self.host = host
        self.port = port
        self.socket = None
        self.connected = False
        self.receive_thread = None
        # Store active connections by room ID
        self.rooms: Dict[str, Set[str]] = {}

    def connect(self):
        """Connect to the TCP server"""
        try:
            self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.socket.connect((self.host, self.port))
            self.connected = True
            print(f"Connected to backend server at {self.host}:{self.port}")
            
            # Start receiving messages in a separate thread
            self.receive_thread = threading.Thread(target=self.receive_messages)
            self.receive_thread.daemon = True
            self.receive_thread.start()
            
            return True
        except Exception as e:
            print(f"Connection error: {repr(e)}")
            return False

    def disconnect(self):
        """Disconnect from the server"""
        if self.socket:
            self.connected = False
            self.socket.close()
            print("Disconnected from backend server")

    def send_message(self, message_type: str, data: dict):
        """Send a JSON message to the server"""
        if not self.connected:
            print("Not connected to the server")
            return False

        message = {
            "type": message_type,
            "request_id": str(uuid.uuid4()),
            "data": data
        }

        try:
            json_msg = json.dumps(message) + "\n"
            self.socket.sendall(json_msg.encode("utf-8"))
            return True
        except Exception as e:
            print(f"Send error: {e}")
            self.connected = False
            return False

    def handle_connection_request(self, request_id: str, headers: dict, url: str, remote_addr: str):
        """Handle new connection request"""
        room_id = url.strip('/')  # Remove leading/trailing slashes
        client_id = str(uuid.uuid4())
        
        print(f"New connection request in room {room_id} from {remote_addr}")
        
        # Initialize room if it doesn't exist
        if room_id not in self.rooms:
            self.rooms[room_id] = set()
        
        # Add client to room
        self.rooms[room_id].add(client_id)
        
        # Send acceptance response
        self.send_message("response", {
            "request_id": request_id,
            "accepted": True,
            "roomID": room_id,
            "clientID": client_id,
            "data": headers
        })

    def handle_new_message(self, message: dict, headers: dict):
        """Handle incoming message"""
        room_id = headers.get("roomID")
        client_id = headers.get("clientID")
        
        if not room_id or not client_id:
            print("Invalid message: missing roomID or clientID")
            return
            
        print(f"Message in room {room_id} from client {client_id}: {message}")
        
        # Broadcast message to all clients in the room
        self.send_message("broadcast", {
            "roomID": room_id,
            "message": message,
            "from": client_id
        })

    def handle_connection_closed(self, headers: dict):
        """Handle connection closure"""
        room_id = headers.get("roomID")
        client_id = headers.get("clientID")
        
        if room_id and client_id:
            if room_id in self.rooms:
                self.rooms[room_id].discard(client_id)
                print(f"Client {client_id} disconnected from room {room_id}")
                
                # Clean up empty rooms
                if not self.rooms[room_id]:
                    del self.rooms[room_id]
                    print(f"Room {room_id} is now empty and removed")

    def receive_messages(self):
        """Receive and process messages from the server"""
        buffer = ""
        try:
            while self.connected:
                data = self.socket.recv(1024)
                if not data:
                    print("Server closed the connection")
                    self.connected = False
                    break

                buffer += data.decode("utf-8")
                while "\n" in buffer:
                    line, buffer = buffer.split("\n", 1)
                    print(f"Received: {line}")
                    if line:
                        try:
                            message = json.loads(line)
                            message_type = message.get("action")
                            data = message.get("params", {})
                            
                            if message_type == "connection_request":
                                self.handle_connection_request(
                                    message.get("request_id"),
                                    data.get("headers", {}),
                                    data.get("url", ""),
                                    data.get("remote_addr", "")
                                )
                            elif message_type == "new_message":
                                self.handle_new_message(
                                    data.get("message", {}),
                                    data.get("headers", {})
                                )
                            elif message_type == "connection_closed":
                                self.handle_connection_closed(
                                    data.get("headers", {})
                                )
                        except json.JSONDecodeError:
                            print(f"Invalid JSON received: {line}")
                        except Exception as e:
                            print(f"Error processing message: {e}")
        except Exception as e:
            if self.connected:  # Only show error if we didn't disconnect intentionally
                print(f"Receive error: {e}")
            self.connected = False

def main():
    """Main function to run the backend client"""
    client = TCPBackendClient()
    
    if not client.connect():
        print("Failed to connect to server")
        return
    
    try:
        # Keep the connection alive
        while client.connected:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\nShutting down...")
    finally:
        client.disconnect()

if __name__ == "__main__":
    main()