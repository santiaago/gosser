import React, { Component } from 'react';
import './App.css';

class App extends Component {
  constructor(props){
    super(props)
    this.state = {
      dots: {}
    }
    var es = new EventSource('http://localhost:8081/api/sse')
    es.onmessage = function(event){
      const msg = JSON.parse(event.data)

      const dots = Object.assign({}, this.state.dots)
      dots[msg.id] = msg
      this.setState({dots: dots})
    }.bind(this)
  }

  componentDidUpdate(prevProps, prevState){
    const ctx = this.refs.canvas.getContext('2d')
    ctx.clearRect(0, 0, 500, 500)
    Object.keys(this.state.dots).forEach(k => {
      const d = this.state.dots[k]
      ctx.fillRect(d.x * 100, d.y * 100, 10, 10)
    })
  }

  render() {
    return (
      <div className='App'>
        <div className='App-header'>
          <h2>Welcome to React</h2>
        </div>
        <p className='App-intro'>
          To get started, edit <code>src/App.js</code> and save to reload.
        </p>
        <canvas ref='canvas' width='500' height='500' />
      </div>
    );
  }
}

export default App;
