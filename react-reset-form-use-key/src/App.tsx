import { useState } from "react";

function App() {
  const [data, setData] = useState({
    id: 0,
    name: "",
    age: 0,
  });
  const [isLoading, setIsLoading] = useState(true);

  const loadData = () => {
    setData({
      id: 1,
      name: "user-1",
      age: 25,
    });
    setIsLoading(false);
  };

  return (
    <div>
      <button type="button" onClick={loadData}>
        load
      </button>
      {isLoading && "loading"}
      {/**
       * key={data.id}によりデータがロードされるとdata.idが変更されFormが再レンダリングされる
       */}
      <Form
        key={data.id}
        data={data}
        onSubmit={(d) => {
          alert(`submit ${JSON.stringify(d)}`);
        }}
        isDisabled={isLoading}
      />
    </div>
  );
}

export default App;

function Form({
  data,
  isDisabled,
  onSubmit,
}: {
  data: { id: number; name: string; age: number };
  isDisabled: boolean;
  onSubmit: (data: { name: string; age: number }) => void;
}) {
  const [name, setName] = useState(data.name);
  const [age, setAge] = useState(data.age);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit({ name, age });
      }}
    >
      <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        disabled={isDisabled}
      />
      <input
        type="number"
        value={age}
        onChange={(e) => setAge(parseInt(e.target.value, 10))}
        disabled={isDisabled}
      />
      <button type="submit" disabled={isDisabled}>
        送信
      </button>
    </form>
  );
}
