import { StrictMode, useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider, useNavigate, useSearchParams } from 'react-router'
import Cookies from "js-cookie";

const router = createBrowserRouter([
  {
    path: "/",
    children: [
      {
        index: true,
        Component: IndexPage,
      },
      {
        path: "login",
        Component: LoginPage
      },
    ],
  },
]);

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)

function IndexPage() {
  const [userInfo, setUserInfo] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate()

  const getUserInfo = async () => {
    setIsLoading(true)
    const res = await fetch("http://localhost:8080/api/me", {
      credentials: "include",
    })

    if (res.status === 401) {
      navigate("/login");
      setIsLoading(false);
      return
    }

    const body = await res.json()
    setUserInfo(body["username"]);
    setIsLoading(false)
  }
  useEffect(() => {
    getUserInfo();
  }, [])
  return (
    <section>
      <h1>UserInfo</h1>
      {isLoading && "loading..."}
      username: {userInfo}
    </section>
  )
}

async function getCSRFToken() {
  await fetch("http://localhost:8080/login", {
    method: "GET",
    credentials: "include",
  })
  return Cookies.get()["_csrf"]
}

function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [csrfToken, setCSRFToken] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [searchParams] = useSearchParams()
  const error = searchParams.get("error")

  useEffect(() => {
    setIsLoading(true)
    getCSRFToken().then((token) => {
      setCSRFToken(token)
      setIsLoading(false)
    })
  }, [])
  return (
    <section>
      <h1>Login</h1>
      {error && <span style={{ color: "red" }}>{error}</span>}
      {isLoading ? "loading..." : (
        <form method="POST" action="http://localhost:8080/login">
          <input type="hidden" name="_csrf" value={csrfToken} />
          <div>
            <label htmlFor="id">ユーザー名</label>
            <input id="username" type="text" name="username" value={username} onChange={(e) => setUsername(e.target.value)} />
          </div>
          <div>
            <label htmlFor="password">パスワード</label>
            <input id="password" type="password" name="password" value={password} onChange={(e) => setPassword(e.target.value)} />
          </div>
          <button type="submit">ログイン</button>
        </form>
      )}
    </section>
  )
}