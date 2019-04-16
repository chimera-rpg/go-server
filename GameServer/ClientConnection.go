package GameServer

import (
  "net"
  "github.com/chimera-rpg/go-common/Net"
  "fmt"
  "log"
  "github.com/chimera-rpg/go-server/GameWorld"
)

// Type ClientConnection provides an object for storing and accessing a
// network connection.
type ClientConnection struct {
  Net.Connection
  id int
  //Player *GameWorld.Player
}

func (client *ClientConnection) GetSocket() (net.Conn) {
  return client.Conn
}

func (client *ClientConnection) GetID() (int) {
  return client.id
}


func NewClientConnection(conn net.Conn, id int) *ClientConnection {
  Net.RegisterCommands()
  cc := ClientConnection{
    id: id,
  }
  cc.SetConn(conn)
  return &cc
}

func (c *ClientConnection) OnExplode(s *GameServer) {
  if r := recover(); r != nil {
    s.RemoveClientByID(c.GetID())
    c.GetSocket().Close()
    log.Print(r.(error))
    log.Print(fmt.Errorf("Client %s(%d) exploded, removing.\n", c.GetSocket().RemoteAddr().String(), c.GetID()))
  }
}

func (c *ClientConnection) HandleHandshake(s *GameServer) {
  c.Send(Net.Command(Net.CommandHandshake{
    Version: Net.VERSION,
    Program: "Chimera Golang Server",
  }))

  hs := c.ReceiveCommandHandshake()

  if hs.Version != Net.VERSION {
      c.Send(Net.Command(Net.CommandBasic{
        Type: Net.NOK,
        String: fmt.Sprintf("Version mismatch, expected %d, got %d", Net.VERSION, hs.Version),
      }))
      panic(fmt.Errorf("Client version mismatch, expected %d, got %d", Net.VERSION, hs.Version))
  }

  c.Send(Net.Command(Net.CommandBasic{
    Type: Net.OK,
    String: "HAY",
  }))
  c.HandleLogin(s)
}

func (c *ClientConnection) HandleLogin(s *GameServer) {
  isWaiting := true
  var cmd Net.Command

  for isWaiting {
    err := c.Receive(&cmd)
    if err != nil {
      panic(fmt.Errorf("Client %s(%d) exploded, removing.\n", c.GetSocket().RemoteAddr().String(), c.GetID()))
    }
    switch t := cmd.(type) {
    case Net.CommandLogin:
      if t.Type == Net.QUERY {
        // TODO: Query if user exists
      } else if t.Type == Net.LOGIN {
        // TODO: Check Database for entry
        if t.User == "nommak" {
          // TODO: Check Database for pass
          if t.Pass == "nommak" {
            c.Send(Net.Command(Net.CommandBasic{
              Type: Net.OK,
              String: "Welcome :)",
            }))
            // Load the Database? Set the player to point to it? dunno
            // Probably Player should be <conn,playerstruct,databaseentry>
            //c.Player = GameWorld.NewOwnerPlayer(c)
            isWaiting = false
          } else {
            c.Send(Net.Command(Net.CommandBasic{
              Type: Net.REJECT,
              String: "Password is wrong",
            }))
          }
        } else {
          c.Send(Net.Command(Net.CommandBasic{
            Type: Net.REJECT,
            String: "Account does not exist",
          }))
        }
      } else if t.Type == Net.REGISTER {
        // TODO: See if User does not exist, send a password confirm to client, then create.
      }
    case Net.CommandBasic:
      if t.Type == Net.CYA {
        s.RemoveClientByID(c.GetID())
        c.GetSocket().Close()
        log.Printf("Client %s(%d) left faithfully.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
        return
      }
    default: // Boot the client if it sends anything else.
      s.RemoveClientByID(c.GetID())
      c.GetSocket().Close()
      log.Printf("Client %s(%d) send bad data, kicking.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
    }
  }
  // If we get to here, then the user has successfully logged in.
  c.HandleCharacterCreation(s)
}

func (c *ClientConnection) HandleCharacterCreation(s *GameServer) {
  var cmd Net.Command
  for {
    err := c.Receive(&cmd)
    if err != nil {
      panic(fmt.Errorf("Client %s(%d) exploded, removing.\n", c.GetSocket().RemoteAddr().String(), c.GetID()))
    }
    switch t := cmd.(type) {
    case Net.CommandBasic:
      if t.Type == Net.CYA {
        s.RemoveClientByID(c.GetID())
        c.GetSocket().Close()
        log.Printf("Client %s(%d) left faithfully.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
        return
      }
    }
  }
  c.HandleGame(s)
}

func (c *ClientConnection) HandleGame(s *GameServer) {
  var cmd Net.Command

  for {
    err := c.Receive(&cmd)
    if err != nil {
      panic(fmt.Errorf("Client %s(%d) exploded, removing.\n", c.GetSocket().RemoteAddr().String(), c.GetID()))
    }
    switch t := cmd.(type) {
    case Net.CommandBasic:
      if t.Type == Net.CYA {
        s.RemoveClientByID(c.GetID())
        c.GetSocket().Close()
        log.Printf("Client %s(%d) left faithfully.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
        return
      }
    }
  }

}

func (c *ClientConnection) HandleTravel(s *GameServer, m *GameWorld.GameMap) {
  //var cmd Net.Command
  // Get list of unique archetype images in the map
  // Send <CRC32>->PNG data for each
}
