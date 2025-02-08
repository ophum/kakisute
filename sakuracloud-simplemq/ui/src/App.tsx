import { useState } from 'react';
import useSWR from 'swr';

interface Job {
  id: string;
  name: string;
  status: string;
  message_id: string;
  created_at: string;
}

export function App() {
  const { data, mutate } = useSWR<Job[]>('http://localhost:8080/jobs')
  const [name, setName] = useState("")

  const createJob = async () => {
    const headers = new Headers
    headers.set("Content-Type", "application/json")
    await fetch("http://localhost:8080/jobs", {
      method: "POST",
      headers: headers,
      body: JSON.stringify({
        name: name
      }),
    })

    mutate()
    setName("")
  }
  return (
    <>
      <input type="text" value={name} onChange={e => setName(e.target.value)} />
      <button type="button" onClick={createJob}>create job</button>
      <table border={1}>
        <thead>
          <tr>
            <th>id</th>
            <th>name</th>
            <th>status</th>
            <th>msgID</th>
            <th>createdAt</th>
          </tr>
        </thead>
        <tbody>
          {data?.map(v => (
            <tr>
              <td>{v.id}</td>
              <td>{v.name}</td>
              <td>{v.status}</td>
              <td>{v.message_id}</td>
              <td>{v.created_at}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  )
}