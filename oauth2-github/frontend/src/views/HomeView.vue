<script setup lang="ts">
import { getRepos, getUserOrgs } from '@/api/api';
import { useAuthStore } from '@/stores/auth.store';
import { useQuery } from "@tanstack/vue-query";
import { ref, watch } from 'vue';
import { RouterLink, useRoute } from 'vue-router';

const route = useRoute();
const authStore = useAuthStore();

if (!authStore.isSignedIn) {
  authStore.redirectSignIn();
}

const user = authStore.user;
const page = ref(parseInt(route.query.page as string || "1"))

const { isPending, isError, isFetching, data, error, refetch } = useQuery({
  queryKey: ["api/repos", page.value],
  queryFn: async () => {
    const repos = await getRepos(page.value)()
    const orgs = await getUserOrgs()
    return {
      repos: repos.repos,
      orgs: orgs.orgs
    }
  }
})
if (isError.value) {
  authStore.redirectSignIn();
}

watch(() => route.query.page, (newPage,) => {
  page.value = parseInt(newPage as string || "1")
  refetch()
})
</script>

<template>
  <main>
    <div v-if="isFetching">Fetching...</div>
    <div v-if="isPending">Loading...</div>
    <div v-else-if="isError && error?.message === 'unauthorized'">認証エラー: 再度ログインしてください。<router-link
        :to="{ name: 'sign-in' }">ログインページ</router-link></div>
    <div v-else-if="data">
      <div style="display: flex; flex-direction: row;">
        <img v-bind:src="user?.avatar_url" width="48px" height="48px" />
        <p>{{ user?.login }}</p>
      </div>
      <table border=1>
        <thead>
          <tr>
            <th>name</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="org in data.orgs" v-bind:key="org.id">
            <td>{{ org.login }}</td>
          </tr>
        </tbody>
      </table>
      <router-link :to="{ name: 'home', query: { page: page - 1 } }">next</router-link>
      {{ page }}
      <router-link :to="{ name: 'home', query: { page: page + 1 } }">next</router-link>
      <table border=1>
        <thead>
          <tr>
            <th>name</th>
            <th>visibility</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="repo in data.repos" v-bind:key="repo.full_name">
            <td>{{ repo.full_name }}</td>
            <td>{{ repo.visibility }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </main>
</template>
