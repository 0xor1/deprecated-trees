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
              <v-card-text>
                <v-form ref="form" @keyup.native.enter="register">
                  <v-text-field v-model="name" prepend-icon="person" name="name" label="Username" type="text"></v-text-field>
                  <v-text-field v-model="displayName" prepend-icon="person" name="displayName" label="Display Name" type="text"></v-text-field>
                  <v-text-field v-model="email" prepend-icon="person" name="email" label="Email" type="email"></v-text-field>
                  <v-text-field v-model="pwd" prepend-icon="lock" name="pwd" label="Password" id="pwd" type="password"></v-text-field>
                  <v-select
                    prepend-icon="location_on"
                    :items="regions"
                    v-model="region"
                    label="Region"
                    single-line
                  ></v-select>
                </v-form>
              </v-card-text>
              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn color="accent" v-on:click="register">Register</v-btn>
              </v-card-actions>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-content>
  </v-app>
</template>

<script>
  import api, {cnst} from '@/api'
  import router from '@/router'
  export default {
    name: 'register',
    data () {
      return {
        name: '',
        displayName: '',
        email: '',
        pwd: '',
        region: null,
        regions: [
          {text: 'US West', value: cnst.regions.usw},
          {text: 'US East', value: cnst.regions.use},
          {text: 'EU West', value: cnst.regions.euw},
          {text: 'Asia Pacific', value: cnst.regions.asp},
          {text: 'Australia', value: cnst.regions.aus}
        ]
      }
    },
    methods: {
      login () {
        router.push('/login')
      },
      register () {
        api.v1.centralAccount.register(this.name, this.email, this.pwd, this.region, 'en', this.displayName, cnst.theme.light).then(() => {
          router.push('/confirmEmail')
        })
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
