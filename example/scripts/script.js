async function test() {
  console.log("JS is working!");
  document.getElementById("result1").innerText = "JS is working!";
  await fetch("https://jsonplaceholder.typicode.com/todos/1")
    .then((response) => response.json())
    .then((json) => {
      console.log(json);
      document.getElementById("result2").innerText = JSON.stringify(
        json,
        null,
        "\t"
      );
    });
  document.getElementById("result1").innerText = "JS is working2!";
}

test();
