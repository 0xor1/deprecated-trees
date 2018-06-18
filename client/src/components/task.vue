<template>
  <v-container class="pa-0" fluid fill-height>
    <v-layout v-if="loading" align-center justify-center>
      <fingerprint-spinner :animation-duration="2000" :size="200" :color="'#405A0F'"></fingerprint-spinner>
    </v-layout>
    <v-layout v-else row>
      <v-layout v-if="projects.length === 0" justify-center align-center>
        <v-card class="mt-3">
          <v-card-text >No Projects</v-card-text> <v-btn v-on:click="toggleCreateForm" color="primary">create</v-btn>
        </v-card>
      </v-layout>
      <v-flex v-else style="overflow-x: auto">
        <v-list  three-line>
          <v-list-tile align-content-start v-for="project in projects" :key="project.id" avatar v-on:click="goToTask(project)">
            <v-list-tile-avatar>
              <v-icon class="green white--text">assignment</v-icon>
            </v-list-tile-avatar>
            <v-list-tile-content style="max-width: 220px;min-width: 220px">
              <v-list-tile-title>{{ project.name }}</v-list-tile-title>
              <v-list-tile-sub-title>{{ project.description }}</v-list-tile-sub-title>
              <v-list-tile-sub-title>Is Public: <h4 style="display: inline;">{{ project.isPublic }}</h4></v-list-tile-sub-title>
            </v-list-tile-content>
            <v-list-tile-content style="max-width: 220px;min-width: 220px">
              <v-container column>
                <v-layout>
                  <v-flex shrink>
                    <v-list-tile-sub-title>Created:&nbsp;</v-list-tile-sub-title>
                    <v-list-tile-sub-title>Start:&nbsp;</v-list-tile-sub-title>
                    <v-list-tile-sub-title>Due:&nbsp;</v-list-tile-sub-title>
                  </v-flex>
                  <v-flex shrink>
                    <v-list-tile-sub-title><h4>{{new Date(project.createdOn).toDateString()}}</h4></v-list-tile-sub-title>
                    <v-list-tile-sub-title><h4>{{project.startOn? new Date(project.startOn).toDateString(): 'No Start Date'}}</h4></v-list-tile-sub-title>
                    <v-list-tile-sub-title><h4>{{project.dueOn? new Date(project.dueOn).toDateString(): 'No Due Date'}}</h4></v-list-tile-sub-title>
                  </v-flex>
                </v-layout>
              </v-container>
            </v-list-tile-content>
            <v-list-tile-content style="max-width: 150px;min-width: 150px">
              <v-container column>
                <v-layout>
                  <v-flex shrink>
                    <v-list-tile-sub-title>Nodes:&nbsp;</v-list-tile-sub-title>
                    <v-list-tile-sub-title>Files:&nbsp;</v-list-tile-sub-title>
                    <v-list-tile-sub-title>File Size:&nbsp;</v-list-tile-sub-title>
                  </v-flex>
                  <v-flex shrink>
                    <v-list-tile-sub-title><h4>{{ project.descendantCount }}</h4></v-list-tile-sub-title>
                    <v-list-tile-sub-title><h4>{{ project.fileCount }}</h4></v-list-tile-sub-title>
                    <v-list-tile-sub-title><h4>{{ project.fileSize }}</h4></v-list-tile-sub-title>
                  </v-flex>
                </v-layout>
              </v-container>
            </v-list-tile-content>
            <v-list-tile-content style="min-width: 250px;">
              <v-container column>
                <v-layout>
                  <v-flex shrink>
                    <v-list-tile-sub-title>Total Remaining Time:&nbsp;</v-list-tile-sub-title>
                    <v-list-tile-sub-title>Minimum Remaining Time:&nbsp;</v-list-tile-sub-title>
                    <v-list-tile-sub-title>Total Logged Time:&nbsp;</v-list-tile-sub-title>
                  </v-flex>
                  <v-flex shrink>
                    <v-list-tile-sub-title><h4>{{ project.totalRemainingTime }}</h4></v-list-tile-sub-title>
                    <v-list-tile-sub-title><h4>{{ project.minimumRemainingTime }}</h4></v-list-tile-sub-title>
                    <v-list-tile-sub-title><h4>{{ project.totalLoggedTime }}</h4></v-list-tile-sub-title>
                  </v-flex>
                  <v-flex shrink>
                    <v-list-tile-sub-title>&nbsp;mins</v-list-tile-sub-title>
                    <v-list-tile-sub-title>&nbsp;mins</v-list-tile-sub-title>
                    <v-list-tile-sub-title>&nbsp;mins</v-list-tile-sub-title>
                  </v-flex>
                </v-layout>
              </v-container>
            </v-list-tile-content>
            <!--<v-list-tile-action column>-->
            <!--<v-flex>-->
            <!--<v-btn icon ripple>-->
            <!--<v-icon color="grey lighten-1">edit</v-icon>-->
            <!--</v-btn>-->
            <!--<v-btn icon ripple class="ml-3">-->
            <!--<v-icon color="grey lighten-1">delete</v-icon>-->
            <!--</v-btn>-->
            <!--</v-flex>-->
            <!--</v-list-tile-action>-->
          </v-list-tile>
        </v-list>
      </v-flex>
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
  export default {
    name: 'task',
    data () {
      return {}
    }
  }
</script>

<style scoped lang="scss">

</style>
