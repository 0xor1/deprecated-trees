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
                <v-btn color="secondary" v-on:click="register">Register</v-btn>
              </v-toolbar>
              <v-card-media src="/static/img/icons/logo.svg" class="mt-3" height=200 contain></v-card-media>
              <v-card-text>
                <v-form ref="form" @keyup.native.enter="login">
                  <v-text-field prepend-icon="person" name="email" label="Email" type="email" v-model="email"></v-text-field>
                  <v-text-field prepend-icon="lock" name="pwd" label="Password" id="pwd" type="password" v-model="pwdTry"></v-text-field>
                </v-form>
              </v-card-text>
              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn color="accent" v-on:click="login">Login</v-btn>
              </v-card-actions>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-content>
  </v-app>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'login',
    data () {
      return {
        email: '',
        pwdTry: ''
      }
    },
    methods: {
      login () {
        api.v1.centralAccount.authenticate(this.email, this.pwdTry).then((res) => {
          let me = res.data.me
          router.push('/app/region/' + me.region + '/shard/' + me.shard + '/account/' + me.id + '/projects')
        }).catch(() => {
          // TODO
        })
      },
      register () {
        router.push('/register')
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
