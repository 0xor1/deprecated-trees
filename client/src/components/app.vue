<template>
  <v-app id="inspire">
    <v-navigation-drawer
      clipped
      fixed
      v-model="drawer"
      app
    >
      <v-list dense>
        <v-list-tile @click="showProjects">
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
      <v-container fluid fill-height>
        <v-layout justify-center align-center>
          <v-flex shrink>
            <v-tooltip right>
              <span>Source</span>
            </v-tooltip>
          </v-flex>
        </v-layout>
      </v-container>
    </v-content>
    <v-footer app fixed>
      <span>&copy; 2017</span>
    </v-footer>
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
      logout: () => {
        api.logout().then(() => {
          router.push('/login')
        })
      },
      showProjects: () => {
        api.v1.centralAccount.getMe().then((res) => {
          let me = res.data.me
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
