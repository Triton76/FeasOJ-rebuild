<script setup>
import { reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { getCaptchaCode, updatePassword } from '../../utils/api/auth';
import { showAlert } from '../../utils/alert';
import { passwordResetEnabled } from '../../utils/account';

const { t } = useI18n();

const networkloading = ref(false);
const formState = reactive({
    email: '',
    vcode: '',
    newPassword: ''
});

const sendCode = async () => {
    if (!passwordResetEnabled.value) {
        return;
    }
    try {
        networkloading.value = true;
        await getCaptchaCode(formState.email, 'false');
        showAlert(t('message.success') + '!', '');
    } catch (error) {
        showAlert(error?.response?.data?.error || t('message.failed') + '!', '');
    } finally {
        networkloading.value = false;
    }
}

const submitReset = async () => {
    if (!passwordResetEnabled.value) {
        return;
    }
    try {
        networkloading.value = true;
        await updatePassword(formState.email, formState.vcode, formState.newPassword);
        showAlert(t('message.success') + '!', '/login');
    } catch (error) {
        showAlert(error?.response?.data?.error || t('message.failed') + '!', '');
    } finally {
        networkloading.value = false;
    }
}
</script>

<template>
    <v-app-bar :elevation="2">
        <template v-slot:prepend>
            <v-btn icon="mdi-chevron-left" size="x-large" @click="$router.back"></v-btn>
            <v-col class="align-left">
                <p style="font-size: 24px;">{{ t("message.resetpwd") }}</p>
            </v-col>
        </template>
    </v-app-bar>
    <v-sheet class="constrainsheet" rounded="xl" :elevation="10">
        <div v-if="!passwordResetEnabled" style="padding: 24px; text-align: center;">
            <v-icon size="64" color="warning">mdi-information-outline</v-icon>
            <p style="margin-top: 16px; font-size: 18px;">
                reset password is disabled in current environment
            </p>
            <v-btn color="primary" rounded="xl" style="margin-top: 16px;" @click="$router.push('/login')">
                {{ $t('message.login') }}
            </v-btn>
        </div>
        <v-form v-else fast-fail width="400px" class="mx-auto" @submit.prevent="submitReset" style="padding: 24px;">
            <v-text-field v-model="formState.email" rounded="xl" variant="solo-filled" :label="$t('message.email')" />
            <v-text-field v-model="formState.vcode" rounded="xl" variant="solo-filled" :label="$t('message.vCode')" />
            <v-text-field v-model="formState.newPassword" rounded="xl" variant="solo-filled" type="password" :label="$t('message.password')" />
            <v-row justify="end">
                <v-btn color="primary" variant="tonal" rounded="xl" @click="sendCode" :loading="networkloading">Send Code</v-btn>
                <div style="margin: 0 8px;"></div>
                <v-btn type="submit" color="primary" rounded="xl" :loading="networkloading">{{ $t('message.resetpwd') }}</v-btn>
            </v-row>
        </v-form>
    </v-sheet>
</template>
