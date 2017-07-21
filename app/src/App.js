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
      console.log(event.data)
    }
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
        <canvas width='500' height='500' />
      </div>
    );
  }
}

export default App;
