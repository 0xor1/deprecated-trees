<template>
    <v-container fluid fill-height>
      <v-layout align-center>
        <v-flex shrink>

        </v-flex>
      </v-layout>
      <v-btn v-on:click="addProject" fixed right bottom color="primary" fab>
        <v-icon>add</v-icon>
      </v-btn>
    </v-container>
</template>

<script>
  export default {
    name: 'projects',
    data () {
      return {
        projects: []
      }
    },
    methods: {
      addProject: () => {
        console.log()
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
