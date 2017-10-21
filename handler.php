<?php

namespace CalProxy;

class handler {

    /**
     * Parse the event and do the replacement and optimizations
     * @param $e array a single ical event that should be cleaned up
     */
    public static function cleanEvent(&$e) {
        //Add missing fields if possible
        if (!isset($e['GEO'])) {
            $e['GEO'] = '';
        }
        if (!isset($e['LOCATION'])) {
            $e['LOCATION'] = '';
        }
        if (!isset($e['LOCATIONTITLE'])) {
            $e['LOCATIONTITLE'] = '';
        }
        if (!isset($e['URL'])) {
            $e['URL'] = '';
        }
        if (!isset($e['DESCRIPTION'])) {
            $e['DESCRIPTION'] = '';
        }

        //Strip added slashes by the parser
        $e['SUMMARY'] = stripcslashes($e['SUMMARY']);
        $e['DESCRIPTION'] = stripcslashes($e['DESCRIPTION']);
        $e['LOCATION'] = stripcslashes($e['LOCATION']);

        //Remember the old title in the description
        $e['DESCRIPTION'] = $e['SUMMARY'] . "\n" . $e['DESCRIPTION'];

        //Remove the TAG and anything after e.g.: (IN0001)
        $e['SUMMARY'] = preg_replace('/(\((IN|MA)[0-9]+,?\s?\)*).+/', '', $e['SUMMARY']);

        //Some common replacements: yes its a long list
        $searchReplace = [];
        $searchReplace['Tutorübungen'] = 'TÜ';
        $searchReplace['Grundlagen'] = 'G';
        $searchReplace['Datenbanken'] = 'DB';
        $searchReplace['Betriebssysteme und Systemsoftware'] = 'BS';
        $searchReplace['Einführung in die Informatik '] = 'INFO';
        $searchReplace['Praktikum: Grundlagen der Programmierung'] = 'PGP';
        $searchReplace['Einführung in die Rechnerarchitektur'] = 'ERA';
        $searchReplace['Einführung in die Softwaretechnik'] = 'EIST';
        $searchReplace['Algorithmen und Datenstrukturen'] = 'AD';
        $searchReplace['Rechnernetze und Verteilte Systeme'] = 'RNVS';
        $searchReplace['Einführung in die Theoretische Informatik'] = 'THEO';
        $searchReplace['Diskrete Strukturen'] = 'DS';
        $searchReplace['Diskrete Wahrscheinlichkeitstheorie'] = 'DWT';
        $searchReplace['Numerisches Programmieren'] = 'NumProg';
        $searchReplace['Lineare Algebra für Informatik'] = 'LinAlg';
        $searchReplace['Analysis für Informatik'] = 'Analysis';
        $searchReplace[' der Künstlichen Intelligenz'] = 'KI';
        $searchReplace['Advanced Topics of Software Engineering'] = 'ASE';
        $searchReplace['Praktikum - iPraktikum, iOS Praktikum'] = 'iPraktikum';

        //Do the replacement
        $e['SUMMARY'] = strtr($e['SUMMARY'], $searchReplace);

        //Remove some stuff which is not really needed
        $e['SUMMARY'] = str_replace(['Standardgruppe', 'PR, ', 'VO, ', 'FA, '], '', $e['SUMMARY']);

        //Try to make sense out of the location
        if (!empty($e['LOCATION'])) {
            if (strpos($e['LOCATION'], '(56') !== false) {
                // Informatik
                switchLocation($e, 'Boltzmannstraße 3, 85748 Garching bei München');
            } else if (strpos($e['LOCATION'], '(55') !== false) {
                // Maschbau
                switchLocation($e, 'Boltzmannstraße 15, 85748 Garching bei München');
            } else if (strpos($e['LOCATION'], '(81') !== false) {
                // Hochbrück
                switchLocation($e, 'Parkring 11-13, 85748 Garching bei München');
            } else if (strpos($e['LOCATION'], '(51') !== false) {
                // Physik
                switchLocation($e, 'James-Franck-Straße 1, 85748 Garching bei München');
            }
        }

        //Check status
        switch ($e['STATUS']) {
            default:
            case 'CONFIRMED':
                $e['STATUS'] = \Eluceo\iCal\Component\Event::STATUS_CONFIRMED;
                break;
            case 'CANCELLED':
                $e['STATUS'] = \Eluceo\iCal\Component\Event::STATUS_CANCELLED;
                break;
            case 'TENTATIVE':
                $e['STATUS'] = \Eluceo\iCal\Component\Event::STATUS_TENTATIVE;
                break;
        }
    }


    /**
     * Update the location field
     *
     * @param $e array element to be edited
     * @param $newLoc string new location that should be set to the element
     */
    public static function switchLocation(&$e, $newLoc) {
        $e['DESCRIPTION'] = $e['LOCATION'] . "\n" . $e['DESCRIPTION'];
        $e['LOCATIONTITLE'] = $e['LOCATION'];
        $e['LOCATION'] = $newLoc;
    }

    /**
     * Remove duplicate entries: events that happen at the same time in multiple locations
     * @param $events
     */
    public static function noDupes(&$events) {
        //Sort them
        usort($events, function ($a, $b) {
            if (strtotime($a['DTSTART']) > strtotime($b['DTSTART'])) {
                return 1;
            } else if ($a['DTSTART'] > $b['DTSTART']) {
                return -1;
            }

            return 0;
        });

        //Find dupes
        $total = count($events);
        for ($i = 1; $i < $total; $i++) {
            //Check if start time, end time and title match then merge
            if ($events[ $i - 1 ]['DTSTART'] === $events[ $i ]['DTSTART']
                && $events[ $i - 1 ]['DTEND'] === $events[ $i ]['DTEND']
                && $events[ $i - 1 ]['SUMMARY'] === $events[ $i ]['SUMMARY']) {
                //Append the location to the next (same) element
                $events[ $i ]['LOCATION'] .= "\n" . $events[ $i - 1 ]['LOCATION'];

                //Mark this element for removal
                unset($events[ $i - 1 ]);
            }
        }
    }
}