<template>
  <v-app id="inspire">
    <v-content>
      <v-container fluid fill-height>
        <v-layout align-center justify-center>
          <v-flex xs12 sm8 md4>
            <v-card class="elevation-12">
              <v-toolbar color="primary">
                <v-toolbar-title>Project Trees</v-toolbar-title>
                <v-spacer></v-spacer>
                <v-btn color="secondary" v-on:click="login">Login</v-btn>
              </v-toolbar>
              <v-card-media src="/static/img/icons/logo.svg" class="mt-3" height=200 contain></v-card-media>
              <v-card-title primary-title>
                <h3 class="headline mb-0">Check Emails For Confirmation Link</h3>
              </v-card-title>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-content>
  </v-app>
</template>

<script>
  import router from '@/router'
  export default {
    name: 'confirmEmail',
    data () {
      return {}
    },
    methods: {
      login () {
        router.push('/login')
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
