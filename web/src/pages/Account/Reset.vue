<!-- 重置密码页（rebuild 阶段已下线） -->
<script setup>
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
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
        <div style="padding: 24px; text-align: center;">
            <v-icon size="64" color="warning">mdi-information-outline</v-icon>
            <p style="margin-top: 16px; font-size: 18px;">
                reset password is disabled in rebuild backend
            </p>
            <v-btn color="primary" rounded="xl" style="margin-top: 16px;" @click="$router.push('/login')">
                {{ $t('message.login') }}
            </v-btn>
        </div>
    </v-sheet>
</template>
