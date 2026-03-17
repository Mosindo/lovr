import React from "react";
import { YStack, type GetProps } from "tamagui";

export type ScreenProps = GetProps<typeof YStack>;

export default function Screen({ children, ...props }: ScreenProps) {
  return (
    <YStack backgroundColor="$background" flex={1} {...props}>
      {children}
    </YStack>
  );
}
