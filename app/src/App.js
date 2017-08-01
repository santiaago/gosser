import React, { Component } from 'react';
import './App.css';

class App extends Component {
  constructor(props){
    super(props)
    this.state = {
      dots: {}
    }
    let es = new EventSource('http://localhost:8081/api/sse')

    es.onmessage = function(event){
      const msg = JSON.parse(event.data)
      console.log('id:', msg.id)
      const dots = Object.assign({}, this.state.dots)
      dots[msg.id] = msg
      this.setState({dots: dots})
    }.bind(this)

    es.addEventListener("newConnection", function(e) {
      var obj = JSON.parse(e.data);
      console.log(obj)
      this.setState({connectionID: obj.id})
    }.bind(this))
  }

  componentDidUpdate(prevProps, prevState){
    const ctx = this.refs.canvas.getContext('2d')
    ctx.clearRect(0, 0, 500, 500)
    Object.keys(this.state.dots).forEach(k => {
      const d = this.state.dots[k]
      if(d.id === this.state.connectionID){
        ctx.fillStyle = 'orange'
        ctx.fillRect(d.x, d.y, 10, 10)
        ctx.fillStyle = 'black'
        return
      }
      ctx.fillRect(d.x, d.y, 10, 10)
    })
  }

  render() {
    return (
      <div className='App'>
        <div className='App-header'>
          <h2>Go server send event demo by Santiago Arias</h2>
        </div>
        <canvas className='world' ref='canvas' height='500' width='500'/>
      </div>
    );
  }
}

export default App;
