export const getRepos = (page: number) => async () => {
  const res = await fetch(`http://localhost:8080/api/repos?page=${page}`, {
    credentials: 'include'
  })
  if (res.status === 401) {
    throw new Error('unauthorzed')
  }

  return res.json()
}

export const getUserOrgs = async () => {
  const res = await fetch(`http://localhost:8080/api/user/orgs`, {
    credentials: 'include'
  })
  if (res.status === 401) {
    throw new Error('unauthorzed')
  }

  return res.json()
}

export const getUser = async () => {
  const res = await fetch('http://localhost:8080/api/user', {
    credentials: 'include'
  })
  if (res.status === 401) {
    throw new Error('unauthorzed')
  }

  return res.json()
}
