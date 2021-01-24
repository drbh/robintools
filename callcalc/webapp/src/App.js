import React, { useEffect } from "react";
import logo from './logo.svg';
import './App.css';


export class App extends React.Component {
  constructor(props) {
    super(props);
    this.data = {}


  }

  componentDidMount() {
    // Create WebSocket connection.
    const socket = new WebSocket('ws://localhost:5000/ws');

    // Connection opened
    socket.addEventListener('open', function (event) {
        socket.send('Hello Server!');
    });

    // Listen for messages
    socket.addEventListener('message', function (event) {
        console.log('Message from server ', event.data);

        var parsed = JSON.parse(event.data)

        if (parsed) {
          var date = parsed["date"]
          

          if (!(date in this.data)) {
            this.data[date] = {}
          }

          var strike = parsed["strike"]
          this.data[date][strike] = JSON.parse(event.data)
          this.forceUpdate()
        }
        
    }.bind(this));
  }

  render() {
    return (
      <div className="App">
      <pre>
        {JSON.stringify(this.data, undefined, 2)}
        </pre>
      </div>
    )
  }
}

export default App;
