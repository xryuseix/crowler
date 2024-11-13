fetch("https://jsonplaceholder.typicode.com/todos/1")
  .then((response) => response.json())
  .then((json) => {
    document.getElementById("result").innerText = JSON.stringify(json, null, "\t");
  });
