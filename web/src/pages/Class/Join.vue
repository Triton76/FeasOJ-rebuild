<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { applyJoinClass } from '../../utils/api/classes';
import { showAlert } from '../../utils/alert';

const router = useRouter();
const classCode = ref('');
const loading = ref(false);

const joinClass = async () => {
    if (!classCode.value.trim()) {
        showAlert('Class code is required', '');
        return;
    }

    loading.value = true;
    try {
        await applyJoinClass(classCode.value.trim());
        showAlert('Join request submitted', '/classes');
    } catch (error) {
        const status = error?.response?.status;
        if (status === 401) {
            showAlert('Please login first', '/login');
            return;
        }
        if (status === 403) {
            showAlert('You do not have permission to join this class', '');
            return;
        }
        if (status === 404) {
            showAlert('Class code not found', '');
            return;
        }
        if (status === 409) {
            showAlert('You already have a membership record for this class', '');
            return;
        }
        showAlert('Join class failed, please retry', '');
    } finally {
        loading.value = false;
    }
};
</script>

<template>
    <v-container fluid class="pa-6">
        <v-row justify="center">
            <v-col cols="12" md="8" lg="6">
                <v-card elevation="2" rounded="lg">
                    <v-card-title class="text-h6">Join Class</v-card-title>
                    <v-divider></v-divider>
                    <v-card-text>
                        <v-text-field
                            v-model="classCode"
                            label="Class Code"
                            variant="solo-filled"
                            density="comfortable"
                            prepend-inner-icon="mdi-key-variant"
                        />
                    </v-card-text>
                    <v-card-actions>
                        <v-btn variant="text" @click="router.push('/classes')">Back</v-btn>
                        <v-spacer></v-spacer>
                        <v-btn color="primary" :loading="loading" @click="joinClass">Submit Request</v-btn>
                    </v-card-actions>
                </v-card>
            </v-col>
        </v-row>
    </v-container>
</template>
