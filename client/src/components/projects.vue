<template>
  <v-container class="pa-0" fluid fill-height>
  <v-layout column fill-height>
    <v-flex>
      <v-text-field
        v-model="search"
        append-icon="search"
        label="Search Name"
        single-line
        hide-details>
      </v-text-field>
    </v-flex>
    <v-data-table
      style="width: 100%!important; height: 100%!important"
      :headers="headers"
      :items="projects"
      :pagination.sync="pagination"
      :total-items="totalProjects"
      :loading="loading"
      hide-actions
      class="elevation-1"
    >
      <template slot="items" slot-scope="projects">
        <tr @click="goToTask(projects.item)" style="cursor: pointer">
          <td class="text-xs-left">{{ projects.item.name }}</td>
          <td class="text-xs-left hidden-sm-and-down">{{ projects.item.description? projects.item.description: 'none' }}</td>
          <td class="text-xs-left" style="width: 150px;">{{ projects.item.startOn? new Date(projects.item.startOn).toLocaleDateString(): 'none' }}</td>
          <td class="text-xs-left" style="width: 150px;">{{ projects.item.dueOn? new Date(projects.item.dueOn).toLocaleDateString(): 'none' }}</td>
          <td class="text-xs-left" style="width: 150px;">{{ projects.item.createdOn? new Date(projects.item.createdOn).toLocaleDateString(): 'none' }}</td>
          <td class="text-xs-left" style="width: 120px;">{{ printDuration(projects.item.minimumRemainingTime, false, projects.item.hoursPerDay, projects.item.daysPerWeek) }}</td>
          <td class="text-xs-left" style="width: 120px;">{{ printDuration(projects.item.totalRemainingTime, false, projects.item.hoursPerDay, projects.item.daysPerWeek) }}</td>
          <td class="text-xs-left" style="width: 120px;">{{ printDuration(projects.item.totalLoggedTime, false, projects.item.hoursPerDay, projects.item.daysPerWeek) }}</td>
        </tr>
      </template>
      <template slot="no-data">
        <v-alert v-if="!loading && !lastUsedSearch" :value="true" color="info" icon="error">
          No Projects Yet <v-btn v-on:click="toggleCreateForm" color="primary">create</v-btn>
        </v-alert>
        <v-alert v-if="!loading && lastUsedSearch" :value="true" color="info" icon="error">
          No Projects Match Search <h3>"{{lastUsedSearch}}"</h3>
        </v-alert>
        <v-alert v-if="loading" :value="true" color="info" icon="error">
          Loading Projects
        </v-alert>
      </template>
    </v-data-table>
      <v-btn v-if="projects.length > 0" v-on:click="toggleCreateForm" fixed right bottom color="primary" fab>
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
          <v-btn icon dark @click.native="toggleCreateForm">
            <v-icon>close</v-icon>
          </v-btn>
        </v-toolbar>
        <v-card-text>
          <v-form ref="form" @keyup.native.enter="createProject" v-model="createProjectValid" lazy-validation>
            <v-text-field v-model="createProjectName" name="projectName" label="Name" type="text"></v-text-field>
            <v-text-field v-model="createProjectDescription" name="projectDescription" label="Description" type="text"></v-text-field>
            <v-layout row>
              <v-flex class="mr-3" xs12 sm6 md4>
                <v-text-field v-model="createProjectHoursPerDay" name="createProjectHoursPerDay" label="Hours per Work Day" type="number" :rules="createProjectHoursPerDayRules"></v-text-field>
              </v-flex>
              <v-flex xs12 sm6 md4>
                <v-text-field v-model="createProjectDaysPerWeek" name="createProjectDaysPerWeek" label="Days per Work Week" type="number" :rules="createProjectDaysPerWeekRules"></v-text-field>
              </v-flex>
            </v-layout>
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
              :loading="creating"
              :disabled="creating || !createProjectValid"
              color="secondary"
              @click="createProject"
            >
              Create
            </v-btn>
          </v-form>
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-layout>
  </v-container>
</template>

<script>
  import api, {cnst} from '@/api'
  import router from '@/router'
  import {printDuration} from '@/helper'
  export default {
    name: 'projects',
    data () {
      return {
        headers: [
          {text: 'Name', sortable: false, align: 'left', value: 'name'},
          {text: 'Description', class: 'hidden-sm-and-down', sortable: false, align: 'left', value: 'description'},
          {text: 'Start', align: 'left', value: cnst.sortBy.startOn},
          {text: 'Due', align: 'left', value: cnst.sortBy.dueOn},
          {text: 'Created', align: 'left', value: cnst.sortBy.createdOn},
          {text: 'Min.', sortable: false, align: 'left', value: 'minimumRemainingTime'},
          {text: 'Tot.', sortable: false, align: 'left', value: 'totalRemainingTime'},
          {text: 'Log.', sortable: false, align: 'left', value: 'totalLoggedTime'}
        ],
        createProjectHoursPerDay: 8,
        createProjectHoursPerDayRules: [
          v => {
            if (!v || v <= 0 || v > 24) {
              return 'must be a positive number no greater than 24'
            }
            return true
          }
        ],
        createProjectDaysPerWeek: 5,
        createProjectDaysPerWeekRules: [
          v => {
            if (!v || v <= 0 || v > 7) {
              return 'must be a positive number no greater than 7'
            }
            return true
          }
        ],
        createProjectValid: true,
        totalProjects: 0,
        pagination: {
          descending: false,
          sortBy: cnst.sortBy.createdOn
        },
        search: '',
        lastUsedSearch: '',
        loading: false,
        createProjectDialog: false,
        createProjectName: '',
        createProjectDescription: null,
        createProjectStartOn: null,
        createProjectStartOnShowPicker: false,
        createProjectDueOn: null,
        createProjectDueOnShowPicker: false,
        createProjectIsParallel: true,
        creating: false,
        projects: [],
        searchSetTimeout: null
      }
    },
    watch: {
      pagination: {
        handler () {
          this.loadProjects(false)
        },
        deep: true
      },
      search (val) {
        clearTimeout(this.searchSetTimeout)
        this.searchSetTimeout = setTimeout(() => {
          this.loadProjects(false)
        }, 500)
      }
    },
    mounted () {
      let htmlEl = document.querySelector('html')
      let self = this
      this.pageScrollListener = () => {
        if (self.totalProjects > self.projects.length && htmlEl.clientHeight + htmlEl.scrollTop >= htmlEl.scrollHeight) {
          self.loadProjects(true)
        }
      }
      document.addEventListener('scroll', this.pageScrollListener)
    },
    beforeDestroy () {
      document.removeEventListener('scroll', this.pageScrollListener)
    },
    methods: {
      printDuration,
      loadProjects (fromScroll) {
        if (!this.loading) {
          let params = router.currentRoute.params
          this.loading = true
          if (!this.pagination.sortBy) {
            this.pagination.sortBy = cnst.sortBy.createdOn
            this.pagination.descending = false
          }
          let after = null
          if (fromScroll && this.projects.length > 0) {
            after = this.projects[this.projects.length - 1].id
          }
          let search = null
          if (this.search && this.search.length > 0) {
            search = this.search
          }
          this.lastUsedSearch = search
          api.v1.project.getSet(params.region, params.shard, params.account, search, null, null, null, null, null, null, false, this.pagination.sortBy, !this.pagination.descending, after, 100).then((res) => {
            this.loading = false
            if (!fromScroll) {
              this.projects = []
            }
            res.projects.forEach((project) => {
              this.projects.push(project)
            })
            this.totalProjects = this.projects.length
            if (res.more) {
              this.totalProjects++
            }
          })
        }
      },
      goToTask (project) {
        let params = router.currentRoute.params
        router.push('/app/region/' + params.region + '/shard/' + params.shard + /account/ + params.account + /project/ + project.id + /task/ + project.id)
      },
      toggleCreateForm () {
        if (!this.creating) {
          this.createProjectDialog = !this.createProjectDialog
          this.createProjectName = ''
          this.createProjectDescription = null
          this.createProjectStartOn = null
          this.createProjectStartOnShowPicker = false
          this.createProjectDueOn = null
          this.createProjectDueOnShowPicker = false
          this.createProjectIsParallel = true
        }
      },
      createProject () {
        if (!this.$refs.form.validate()) {
          return
        }
        this.creating = true
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
        api.v1.project.create(params.region, params.shard, params.account, this.createProjectName, description, parseInt(this.createProjectHoursPerDay), parseInt(this.createProjectDaysPerWeek), startOn, dueOn, this.createProjectIsParallel, false).then((newProject) => {
          this.creating = false
          this.toggleCreateForm()
          let params = router.currentRoute.params
          router.push('/app/region/' + params.region + '/shard/' + params.shard + /account/ + params.account + /project/ + newProject.id + /task/ + newProject.id)
        })
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
