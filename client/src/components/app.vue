<template>
  <v-app id="inspire">
    <v-navigation-drawer
      clipped
      fixed
      v-model="drawer"
      app
    >
      <v-list dense>
        <v-list-tile v-on:click="showProjects">
          <v-list-tile-action>
            <v-icon>group_work</v-icon>
          </v-list-tile-action>
          <v-list-tile-content>
            <v-list-tile-title>Projects</v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
        <v-list-tile @click="logout">
          <v-list-tile-action>
            <v-icon>exit_to_app</v-icon>
          </v-list-tile-action>
          <v-list-tile-content>
            <v-list-tile-title>Logout</v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
      </v-list>
    </v-navigation-drawer>
    <v-toolbar app fixed clipped-left color="primary">
      <v-toolbar-side-icon @click.stop="drawer = !drawer"></v-toolbar-side-icon>
      <v-toolbar-title>Project Trees</v-toolbar-title>
    </v-toolbar>
    <v-content>
      <router-view></router-view>
    </v-content>
  </v-app>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'app',
    data () {
      return {
        drawer: false
      }
    },
    methods: {
      logout () {
        api.logout().then(() => {
          router.push('/login')
        })
      },
      showProjects () {
        api.v1.centralAccount.getMe().then((me) => {
          router.push('/app/region/' + me.region + '/shard/' + me.shard + '/account/' + me.id + '/projects')
        }).catch(() => {
          // TODO
        })
      }
    }
  }
</script>

<style lang="scss">

</style>
