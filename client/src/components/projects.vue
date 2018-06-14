<template>
  <v-container class="pa-0" fluid fill-height>
    <v-layout v-if="loadingProjects" align-center justify-center>
      <fingerprint-spinner :animation-duration="2000" :size="200" :color="'#405A0F'"></fingerprint-spinner>
    </v-layout>
  <v-layout v-else row>
    <v-layout v-if="projects.length === 0" justify-center align-center>
      <v-card class="mt-3">
        <v-card-text >No Projects</v-card-text> <v-btn v-on:click="createProjectDialog = true" color="primary">create</v-btn>
      </v-card>
    </v-layout>
      <v-flex v-else>
      <v-list  two-line>
        <v-list-tile v-for="project in projects" :key="project.id" v-on:click="goToTask(project)">
          <v-list-tile-title>{{project.name}}</v-list-tile-title>
          <v-list-tile-action-text>{{project.description}}</v-list-tile-action-text>
        </v-list-tile>
      </v-list>
      </v-flex>
      <v-btn v-if="projects.length > 0" v-on:click="createProjectDialog = true" fixed right bottom color="primary" fab>
        <v-icon>add</v-icon>
      </v-btn>
    <v-dialog
      v-model="createProjectDialog"
      fullscreen
      hide-overlay
      transition="dialog-bottom-transition"
      scrollable
    >
      <v-card tile>
        <v-toolbar card dark color="primary">
          <v-toolbar-title>Create Project</v-toolbar-title>
          <v-spacer></v-spacer>
          <v-btn icon dark @click.native="createProjectDialog = false">
            <v-icon>close</v-icon>
          </v-btn>
        </v-toolbar>
        <v-card-text>
          <v-form ref="form" @keyup.native.enter="register">
            <v-text-field v-model="createProjectName" name="projectName" label="Name" type="text"></v-text-field>
            <v-text-field v-model="createProjectDescription" name="projectDescription" label="Description" type="text"></v-text-field>
            <v-layout row>
            <v-flex class="mr-3" xs12 sm6 md4>
              <v-menu
                ref="createProjectStartOn"
                :close-on-content-click="false"
                v-model="createProjectStartOnShowPicker"
                :nudge-right="40"
                :return-value.sync="createProjectStartOn"
                lazy
                transition="scale-transition"
                offset-y
                full-width
                min-width="290px"
              >
                <v-text-field
                  slot="activator"
                  v-model="createProjectStartOn"
                  label="Start On"
                  prepend-icon="event"
                  readonly
                ></v-text-field>
                <v-date-picker v-model="createProjectStartOn" @input="$refs.createProjectStartOn.save(createProjectStartOn)"></v-date-picker>

              </v-menu>
            </v-flex>
            <v-flex xs12 sm6 md4>
              <v-menu
                ref="createProjectDueOn"
                :close-on-content-click="false"
                v-model="createProjectDueOnShowPicker"
                :nudge-right="40"
                :return-value.sync="createProjectDueOn"
                lazy
                transition="scale-transition"
                offset-y
                full-width
                min-width="290px"
              >
                <v-text-field
                  slot="activator"
                  v-model="createProjectDueOn"
                  label="Due On"
                  prepend-icon="event"
                  readonly
                ></v-text-field>
                <v-date-picker v-model="createProjectDueOn" @input="$refs.createProjectDueOn.save(createProjectDueOn)"></v-date-picker>
              </v-menu>
            </v-flex>
            </v-layout>
            <v-switch
              :label="`Is Parallel`"
              v-model="createProjectIsParallel"></v-switch>
            <v-btn
              :loading="creatingProject"
              :disabled="creatingProject"
              color="secondary"
              @click="createProject"
            >
              Create
            </v-btn>
          </v-form>
        </v-card-text>

        <div style="flex: 1 1 auto;"></div>
      </v-card>
    </v-dialog>
  </v-layout>
  </v-container>
</template>

<script>
  import api, {cnst} from '@/api'
  import router from '@/router'
  import {FingerprintSpinner} from 'epic-spinners'
  export default {
    name: 'projects',
    components: {FingerprintSpinner},
    data () {
      let params = router.currentRoute.params
      let data = {
        loadingProjects: true,
        createProjectDialog: false,
        createProjectName: '',
        createProjectDescription: null,
        createProjectStartOn: null,
        createProjectStartOnShowPicker: false,
        createProjectDueOn: null,
        createProjectDueOnShowPicker: false,
        createProjectIsParallel: true,
        creatingProject: false,
        projects: []
      }
      api.v1.project.getSet(params.region, params.shard, params.account, null, null, null, null, null, null, null, false, cnst.sortBy.createdOn, cnst.sortDir.asc, null, 100).then((res) => {
        data.loadingProjects = false
        data.projects = res.projects
      })
      return data
    },
    methods: {
      goToTask (project) {
        let params = router.currentRoute.params
        router.push('/app/region/' + params.region + '/shard/' + params.shard + /account/ + params.account + /project/ + project.id + /task/ + project.id)
      },
      createProject () {
        this.creatingProject = true
        let params = router.currentRoute.params
        let description = null
        if (this.createProjectDescription && this.createProjectDescription.length) {
          description = this.createProjectDescription
        }
        let startOn = null
        if (this.createProjectStartOn && this.createProjectStartOn.length > 0) {
          startOn = this.createProjectStartOn + 'T00:00:00Z'
        }
        let dueOn = null
        if (this.createProjectDueOn && this.createProjectDueOn.length > 0) {
          dueOn = this.createProjectDueOn + 'T00:00:00Z'
        }
        api.v1.project.create(params.region, params.shard, params.account, this.createProjectName, description, startOn, dueOn, this.createProjectIsParallel, false).then((newProject) => {
          this.creatingProject = false
          this.createProjectDialog = false
          this.projects.push(newProject)
        })
      }
    }
  }
</script>

<style scoped lang="scss">
.projects-list {
  width: 100%;
}
</style>
