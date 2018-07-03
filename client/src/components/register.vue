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
                <v-form ref="form" @keyup.native.enter="register" v-model="valid" lazy-validation>
                  <v-text-field v-model="name" prepend-icon="person" name="name" label="Username" type="text" :rules="nameRules" required></v-text-field>
                  <v-text-field v-model="displayName" prepend-icon="person" name="displayName" label="Display Name" type="text"></v-text-field>
                  <v-text-field v-model="email" prepend-icon="person" name="email" label="Email" type="email" :rules="emailRules" required></v-text-field>
                  <v-text-field v-model="pwd" prepend-icon="lock" name="pwd" label="Password" id="pwd" type="password" :rules="pwdRules" required></v-text-field>
                  <v-select
                    prepend-icon="location_on"
                    :items="regions"
                    v-model="region"
                    label="Region"
                    single-line
                    :rules="regionRules"
                    required
                  ></v-select>
                </v-form>
              </v-card-text>
              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn color="accent" v-on:click="register" :disabled="!valid">Register</v-btn>
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
        valid: true,
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
        ],
        nameRules: [
          v => {
            if (!v || v.length < 3 || v.length > 50 || !/^\w{3,50}$/.test(v)) {
              return 'Name must be 3-50 numbers, letters or underscore characters'
            }
            return true
          }
        ],
        emailRules: [
          v => {
            if (!v || v.length < 3 || v.length > 50 || !/.+@.+\..+/.test(v)) {
              return 'Valid email required'
            }
            return true
          }
        ],
        pwdRules: [
          v => {
            if (!v || v.length < 8 || v.length > 200 || !/[0-9]/.test(v) || !/[a-z]/.test(v) || !/[A-Z]/.test(v) || !/[\W]/.test(v)) {
              return 'Password must be 8 or more characters including a digit, an upper and lowercase letter and a symbol'
            }
            return true
          }
        ],
        regionRules: [
          v => {
            if (!v || v.length < 3) {
              return 'Must select a region'
            }
            return true
          }
        ]
      }
    },
    methods: {
      login () {
        router.push('/login')
      },
      register () {
        if (this.$refs.form.validate()) {
          api.v1.centralAccount.register(this.name, this.email, this.pwd, this.region, 'en', this.displayName, cnst.theme.light).then(() => {
            router.push('/confirmEmail')
          })
        }
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
